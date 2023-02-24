package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestAuthRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:  "my-state",
					Nonce:  "my-nonce",
					Locale: "en",
				},
			},
		}, nil)

	AuthRedirect(nil, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.LoginCallback+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
}

func TestAuthRedirectWithIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Identity: true,
					LpaID:    "123",
				},
			},
		}, nil)

	AuthRedirect(nil, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/123"+Paths.IdentityWithOneLoginCallback+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
}

func TestAuthRedirectWithCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:               "my-state",
					Nonce:               "my-nonce",
					Locale:              "en",
					Identity:            true,
					CertificateProvider: true,
					SessionID:           "456",
					LpaID:               "123",
				},
			},
		}, nil)

	AuthRedirect(nil, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.CertificateProviderLoginCallback+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
}

func TestAuthRedirectWithCyLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:  "my-state",
					Nonce:  "my-nonce",
					Locale: "cy",
				},
			},
		}, nil)

	AuthRedirect(nil, sessionStore)(w, r)
	resp := w.Result()

	redirect := fmt.Sprintf("/cy%s?code=auth-code&state=my-state", Paths.LoginCallback)

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, redirect, resp.Header.Get("Location"))
}

func TestAuthRedirectSessionMissing(t *testing.T) {
	testCases := map[string]struct {
		url         string
		session     *sessions.Session
		getErr      error
		expectedErr interface{}
	}{
		"missing session": {
			url:         "/?code=auth-code&state=my-state",
			session:     nil,
			getErr:      ExpectedError,
			expectedErr: ExpectedError,
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

			logger := newMockLogger(t)
			logger.
				On("Print", tc.expectedErr)

			sessionStore := newMockSessionStore(t)
			sessionStore.
				On("Get", r, "params").
				Return(tc.session, tc.getErr)

			AuthRedirect(logger, sessionStore)(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestAuthRedirectStateIncorrect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hello", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", "state incorrect")

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce"},
			},
		}, nil)

	AuthRedirect(logger, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
