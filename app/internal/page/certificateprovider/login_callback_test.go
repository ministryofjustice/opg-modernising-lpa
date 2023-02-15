package certificateprovider

import (
	"context"
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

func TestGetLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{Sub: "a-sub", Email: "a-email", CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "John Doe", RetrievedAt: now}

	sessionStore := &MockSessionsStore{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", ctxMatcher).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", ctxMatcher, &page.Lpa{
			CertificateProviderUserData: userData,
		}).
		Return(nil)

	oneLoginClient := &MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", ctxMatcher, "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", ctxMatcher, "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", ctxMatcher, userInfo).
		Return(userData, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &loginCallbackData{
			App:         TestAppData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(TestAppData, w, r)
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
			error:    ExpectedError,
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

			lpaStore := &MockLpaStore{}
			lpaStore.
				On("Get", mock.Anything).
				Return(&page.Lpa{}, nil)

			sessionStore := &MockSessionsStore{}
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

			oneLoginClient := &MockOneLoginClient{}
			oneLoginClient.
				On("Exchange", mock.Anything, mock.Anything, mock.Anything).
				Return("a-jwt", nil)
			oneLoginClient.
				On("UserInfo", mock.Anything, mock.Anything).
				Return(userInfo, nil)
			oneLoginClient.
				On("ParseIdentityClaim", mock.Anything, mock.Anything).
				Return(tc.userData, tc.error)

			template := &MockTemplate{}
			template.
				On("Func", w, &loginCallbackData{
					App:             TestAppData,
					CouldNotConfirm: true,
				}).
				Return(nil)

			err := LoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(TestAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetLoginCallbackWhenExchangeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	sessionStore := &MockSessionsStore{}
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

	oneLoginClient := &MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("", ExpectedError)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	sessionStore := &MockSessionsStore{}
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

	oneLoginClient := &MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{}, ExpectedError)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	sessionStore := &MockSessionsStore{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, ExpectedError)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore)
}

func TestGetLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, mock.Anything).
		Return(ExpectedError)

	sessionStore := &MockSessionsStore{}
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

	oneLoginClient := &MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	sessionStore := &MockSessionsStore{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{CertificateProviderUserData: userData}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &loginCallbackData{
			App:         TestAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Func, nil, sessionStore, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionStore, lpaStore, template)
}

func TestPostCertificateProviderLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := &MockSessionsStore{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
		})).
		Return(&page.Lpa{CertificateProviderUserData: identity.UserData{OK: true}}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderYourDetails, resp.Header.Get("Location"))
}

func TestPostCertificateProviderLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := &MockSessionsStore{}
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

	lpaStore := &MockLpaStore{}
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}
