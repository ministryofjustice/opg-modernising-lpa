package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionsStore := &mockSessionsStore{}

	session := sessions.NewSession(sessionsStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"donor": &sesh.DonorSession{Sub: "random", Email: "name@example.com"},
	}

	sessionsStore.
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
	sessionsStore.
		On("Save", r, w, session).
		Return(nil)

	AuthRedirect(nil, client, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.Dashboard, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestAuthRedirectWithIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionsStore := &mockSessionsStore{}

	sessionsStore.
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

	AuthRedirect(nil, nil, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/123"+Paths.IdentityWithOneLoginCallback+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestAuthRedirectWithCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionsStore := &mockSessionsStore{}

	sessionsStore.
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

	AuthRedirect(nil, nil, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.CertificateProviderLoginCallback+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestAuthRedirectWithCyLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionsStore := &mockSessionsStore{}

	session := sessions.NewSession(sessionsStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{
		"donor": &sesh.DonorSession{Sub: "random", Email: "name@example.com"},
	}

	sessionsStore.
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
	sessionsStore.
		On("Save", r, w, session).
		Return(nil)

	AuthRedirect(nil, client, sessionsStore)(w, r)
	resp := w.Result()

	redirect := fmt.Sprintf("/cy%s", Paths.Dashboard)

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, redirect, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
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

			sessionsStore := &mockSessionsStore{}
			sessionsStore.
				On("Get", r, "params").
				Return(tc.session, tc.getErr)

			AuthRedirect(logger, nil, sessionsStore)(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, sessionsStore)
		})
	}
}

func TestAuthRedirectStateIncorrect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hello", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", "state incorrect")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce"},
			},
		}, nil)

	AuthRedirect(logger, nil, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestAuthRedirectWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", ExpectedError)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("", ExpectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en"},
			},
		}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)
}

func TestAuthRedirectWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", ExpectedError)

	client := newMockOneLoginClient(t)
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{}, ExpectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en"},
			},
		}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)
}
