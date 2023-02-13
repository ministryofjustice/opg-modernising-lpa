package page

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
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

func TestAuthRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := &mockOneLoginClient{}
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
	mock.AssertExpectationsForObjects(t, client, sessionsStore)
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

	client := &mockOneLoginClient{}
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
	mock.AssertExpectationsForObjects(t, client, sessionsStore)
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

			logger := &mockLogger{}
			logger.
				On("Print", tc.expectedErr)

			sessionsStore := &mockSessionsStore{}
			sessionsStore.
				On("Get", r, "params").
				Return(tc.session, tc.getErr)

			AuthRedirect(logger, nil, sessionsStore)(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, logger, sessionsStore)
		})
	}
}

func TestAuthRedirectStateIncorrect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hello", nil)

	logger := &mockLogger{}
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
	mock.AssertExpectationsForObjects(t, logger, sessionsStore)
}

func TestAuthRedirectWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockOneLoginClient{}
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("", expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en"},
			},
		}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}

func TestAuthRedirectWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockOneLoginClient{}
	client.
		On("Exchange", r.Context(), "auth-code", "my-nonce").
		Return("a JWT", nil)
	client.
		On("UserInfo", r.Context(), "a JWT").
		Return(onelogin.UserInfo{}, expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en"},
			},
		}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}
