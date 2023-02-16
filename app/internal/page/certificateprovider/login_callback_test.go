package certificateprovider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockOneLoginClient struct {
	mock.Mock
}

func (m *mockOneLoginClient) AuthCodeURL(state, nonce, locale string, identity bool) string {
	args := m.Called(state, nonce, locale, identity)
	return args.String(0)
}

func (m *mockOneLoginClient) Exchange(ctx context.Context, code, nonce string) (string, error) {
	args := m.Called(ctx, code, nonce)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockOneLoginClient) UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(onelogin.UserInfo), args.Error(1)
}

func (m *mockOneLoginClient) ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error) {
	args := m.Called(ctx, userInfo)
	return args.Get(0).(identity.UserData), args.Error(1)
}

type mockLpaStore struct {
	mock.Mock
}

func (m *mockLpaStore) Create(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)

	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Get(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Put(ctx context.Context, v *page.Lpa) error {
	return m.Called(ctx, v).Error(0)
}

func TestGetLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{Sub: "a-sub", Email: "a-email", CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "John Doe", RetrievedAt: now}

	sessionStore := &mockSessionsStore{}
	session := sessions.NewSession(sessionStore, "session")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"certificate-provider": &sesh.CertificateProviderSession{
			Sub:            "a-sub",
			Email:          "a-email",
			LpaID:          "lpa-id",
			DonorSessionID: "session-id",
		},
	}

	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	ctxMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		session := page.SessionDataFromContext(ctx)

		return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
	})

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", ctxMatcher).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", ctxMatcher, &page.Lpa{
			CertificateProviderUserData: userData,
		}).
		Return(nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", ctxMatcher, "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", ctxMatcher, "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", ctxMatcher, userInfo).
		Return(userData, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &loginCallbackData{
			App:         appData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient, template)
}

func TestGetLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	testCases := map[string]struct {
		userData identity.UserData
		url      string
		error    error
	}{
		"not ok": {
			url: "/?code=a-code",
		},
		"errored": {
			url:      "/?code=a-code",
			userData: identity.UserData{OK: true},
			error:    expectedError,
		},
		"provider access denied": {
			url:      "/?error=access_denied",
			userData: identity.UserData{OK: true},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything).
				Return(&page.Lpa{}, nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{
							State:               "a-state",
							Nonce:               "a-nonce",
							CertificateProvider: true,
							Identity:            true,
							LpaID:               "lpa-id",
							SessionID:           "session-id",
						},
					},
				}, nil)

			oneLoginClient := &mockOneLoginClient{}
			oneLoginClient.
				On("Exchange", mock.Anything, mock.Anything, mock.Anything).
				Return("a-jwt", nil)
			oneLoginClient.
				On("UserInfo", mock.Anything, mock.Anything).
				Return(userInfo, nil)
			oneLoginClient.
				On("ParseIdentityClaim", mock.Anything, mock.Anything).
				Return(tc.userData, tc.error)

			template := &mockTemplate{}
			template.
				On("Func", w, &loginCallbackData{
					App:             appData,
					CouldNotConfirm: true,
				}).
				Return(nil)

			err := LoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetLoginCallbackWhenExchangeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("", expectedError)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{}, expectedError)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, expectedError)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore)
}

func TestGetLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, mock.Anything).
		Return(expectedError)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userInfo := onelogin.UserInfo{Sub: "a-sub", Email: "a-email", CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)

	sessionStore := &mockSessionsStore{}
	session := sessions.NewSession(sessionStore, "session")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"certificate-provider": &sesh.CertificateProviderSession{
			Sub:            "a-sub",
			Email:          "a-email",
			LpaID:          "lpa-id",
			DonorSessionID: "session-id",
		},
	}

	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "a-state",
					Nonce:               "a-nonce",
					CertificateProvider: true,
					Identity:            true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{CertificateProviderUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &loginCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore, template)
}

func TestPostCertificateProviderLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"certificate-provider": &sesh.CertificateProviderSession{
					Sub:            "xyz",
					LpaID:          "lpa-id",
					DonorSessionID: "session-id",
				},
			},
		}, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
		})).
		Return(&page.Lpa{CertificateProviderUserData: identity.UserData{OK: true}}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderYourDetails, resp.Header.Get("Location"))
}

func TestPostCertificateProviderLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"certificate-provider": &sesh.CertificateProviderSession{
					Sub:            "xyz",
					LpaID:          "lpa-id",
					DonorSessionID: "session-id",
				},
			},
		}, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}
