package certificateprovider

import (
	"context"
	io "io"
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

	sessionStore := newMockSessionStore(t)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", ctxMatcher).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", ctxMatcher, &page.Lpa{
			CertificateProviderUserData: userData,
		}).
		Return(nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", ctxMatcher, "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", ctxMatcher, "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", ctxMatcher, userInfo).
		Return(userData, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &loginCallbackData{
			App:         testAppData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Execute, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	testCases := map[string]struct {
		oneLoginClient func(*testing.T) *mockOneLoginClient
		template       func(*testing.T, io.Writer) *mockTemplate
		url            string
		error          error
	}{
		"not ok": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{}, nil)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				template := newMockTemplate(t)
				template.
					On("Execute", w, &loginCallbackData{
						App:             testAppData,
						CouldNotConfirm: true,
					}).
					Return(nil)
				return template
			},
		},
		"errored on parse": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{OK: true}, expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},
			error: expectedError,
		},
		"errored on userinfo": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(onelogin.UserInfo{}, expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},
			error: expectedError,
		},
		"errored on exchange": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("", expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},
			error: expectedError,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return nil
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				template := newMockTemplate(t)
				template.
					On("Execute", w, &loginCallbackData{
						App:             testAppData,
						CouldNotConfirm: true,
					}).
					Return(nil)
				return template
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", mock.Anything).
				Return(&page.Lpa{}, nil)

			sessionStore := newMockSessionStore(t)
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

			oneLoginClient := tc.oneLoginClient(t)
			template := tc.template(t, w)

			err := LoginCallback(template.Execute, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	sessionStore := newMockSessionStore(t)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, expectedError)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, mock.Anything).
		Return(expectedError)

	sessionStore := newMockSessionStore(t)
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

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := LoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userInfo := onelogin.UserInfo{Sub: "a-sub", Email: "a-email", CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)

	sessionStore := newMockSessionStore(t)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{CertificateProviderUserData: userData}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &loginCallbackData{
			App:         testAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := LoginCallback(template.Execute, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := newMockSessionStore(t)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
		})).
		Return(&page.Lpa{CertificateProviderUserData: identity.UserData{OK: true}}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderYourDetails, resp.Header.Get("Location"))
}

func TestPostCertificateProviderLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := newMockSessionStore(t)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", mock.Anything).Return(&page.Lpa{}, nil)

	err := LoginCallback(nil, nil, sessionStore, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}
