package certificateprovider

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

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
			IDToken: "id-token",
			Sub:     "random",
			Email:   "name@example.com",
			LpaID:   "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "my-state",
					Nonce:               "my-nonce",
					Locale:              "en",
					CertificateProvider: true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: base64.StdEncoding.EncodeToString([]byte("random")),
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Create", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	err := LoginCallback(client, sessionStore, certificateProviderStore)(testAppData, w, r)
	assert.Nil(t, err)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderEnterDateOfBirth, resp.Header.Get("Location"))
}

func TestLoginCallbackWhenCertificateProviderExists(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

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
			IDToken: "id-token",
			Sub:     "random",
			Email:   "name@example.com",
			LpaID:   "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "my-state",
					Nonce:               "my-nonce",
					Locale:              "en",
					CertificateProvider: true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: base64.StdEncoding.EncodeToString([]byte("random")),
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Create", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, &types.ConditionalCheckFailedException{})

	err := LoginCallback(client, sessionStore, certificateProviderStore)(testAppData, w, r)
	assert.Nil(t, err)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderEnterDateOfBirth, resp.Header.Get("Location"))
}

func TestLoginCallbackSessionMissing(t *testing.T) {
	testCases := map[string]struct {
		url         string
		session     *sessions.Session
		getErr      error
		expectedErr interface{}
	}{
		"missing session": {
			url:         "/?code=auth-code&state=my-state",
			session:     nil,
			getErr:      expectedError,
			expectedErr: expectedError,
		},
		"missing state": {
			url:         "/?code=auth-code&state=my-state",
			session:     &sessions.Session{Values: map[any]any{}},
			expectedErr: sesh.MissingSessionError("one-login"),
		},
		"missing state from url": {
			url: "/?code=auth-code",
			session: &sessions.Session{
				Values: map[any]any{
					"one-login": &sesh.OneLoginSession{State: "my-state"},
				},
			},
			expectedErr: sesh.InvalidSessionError("one-login"),
		},
		"missing nonce": {
			url: "/?code=auth-code&state=my-state",
			session: &sessions.Session{
				Values: map[any]any{
					"one-login": &sesh.OneLoginSession{State: "my-state", Locale: "en"},
				},
			},
			expectedErr: sesh.InvalidSessionError("one-login"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.
				On("Get", r, "params").
				Return(tc.session, tc.getErr)

			err := LoginCallback(nil, sessionStore, nil)(testAppData, w, r)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestLoginCallbackWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("", "", expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", CertificateProvider: true, LpaID: "lpa-id", SessionID: "session-id"},
			},
		}, nil)

	err := LoginCallback(client, sessionStore, nil)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", CertificateProvider: true, LpaID: "lpa-id", SessionID: "session-id"},
			},
		}, nil)

	err := LoginCallback(client, sessionStore, nil)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackOnCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

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
			IDToken: "id-token",
			Sub:     "random",
			Email:   "name@example.com",
			LpaID:   "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "my-state",
					Nonce:               "my-nonce",
					Locale:              "en",
					CertificateProvider: true,
					LpaID:               "lpa-id",
					SessionID:           "session-id",
				},
			},
		}, nil)
	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: base64.StdEncoding.EncodeToString([]byte("random")),
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Create", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := LoginCallback(client, sessionStore, certificateProviderStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
