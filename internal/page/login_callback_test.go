package page

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginCallback(t *testing.T) {
	testCases := map[string]struct {
		subExists        bool
		expectedRedirect Path
	}{
		"Sub exists":         {subExists: true, expectedRedirect: Paths.Dashboard},
		"Sub does not exist": {subExists: false, expectedRedirect: Paths.Attorney.EnterReferenceNumber},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

			client := newMockOneLoginClient(t)
			client.EXPECT().
				Exchange(r.Context(), "auth-code", "my-nonce").
				Return("id-token", "a JWT", nil)
			client.EXPECT().
				UserInfo(r.Context(), "a JWT").
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
				"session": &sesh.LoginSession{
					IDToken: "id-token",
					Sub:     "random",
					Email:   "name@example.com",
				},
			}

			sessionStore.EXPECT().
				Get(r, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{
							State:    "my-state",
							Nonce:    "my-nonce",
							Locale:   "en",
							Redirect: "/redirect",
						},
					},
				}, nil)
			sessionStore.EXPECT().
				Save(r, w, session).
				Return(nil)

			dashboardStore := newMockDashboardStore(t)
			dashboardStore.EXPECT().
				SubExistsForActorType(r.Context(), base64.StdEncoding.EncodeToString([]byte("random")), actor.TypeAttorney).
				Return(tc.subExists, nil)

			err := LoginCallback(client, sessionStore, Paths.Attorney.EnterReferenceNumber, dashboardStore, actor.TypeAttorney)(AppData{}, w, r)
			assert.Nil(t, err)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect.Format(), resp.Header.Get("Location"))
		})
	}
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
			sessionStore.EXPECT().
				Get(r, "params").
				Return(tc.session, tc.getErr)

			err := LoginCallback(nil, sessionStore, Paths.Attorney.LoginCallback, nil, actor.TypeAttorney)(AppData{}, w, r)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestLoginCallbackWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("", "", expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: Paths.LoginCallback.Format()},
			},
		}, nil)

	err := LoginCallback(client, sessionStore, Paths.LoginCallback, nil, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: Paths.LoginCallback.Format()},
			},
		}, nil)

	err := LoginCallback(client, sessionStore, Paths.LoginCallback, nil, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: Paths.LoginCallback.Format(),
				},
			},
		}, nil)
	sessionStore.EXPECT().
		Save(r, w, mock.Anything).
		Return(expectedError)

	err := LoginCallback(client, sessionStore, Paths.LoginCallback, nil, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenDashboardStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: Paths.LoginCallback.Format(),
				},
			},
		}, nil)
	sessionStore.EXPECT().
		Save(r, w, mock.Anything).
		Return(nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		SubExistsForActorType(r.Context(), mock.Anything, mock.Anything).
		Return(false, expectedError)

	err := LoginCallback(client, sessionStore, Paths.LoginCallback, dashboardStore, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}
