package page

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

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
			sessionStore.EXPECT().
				OneLogin(r).
				Return(&sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				}, nil)
			sessionStore.EXPECT().
				SetLogin(r, w, &sesh.LoginSession{
					IDToken: "id-token",
					Sub:     "random",
					Email:   "name@example.com",
				}).
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

func TestLoginCallbackSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(nil, expectedError)

	err := LoginCallback(nil, sessionStore, Paths.Attorney.LoginCallback, nil, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
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
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: Paths.LoginCallback.Format()}, nil)

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
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: Paths.LoginCallback.Format()}, nil)

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
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: Paths.LoginCallback.Format(),
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
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
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: Paths.LoginCallback.Format(),
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		SubExistsForActorType(r.Context(), mock.Anything, mock.Anything).
		Return(false, expectedError)

	err := LoginCallback(client, sessionStore, Paths.LoginCallback, dashboardStore, actor.TypeAttorney)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}
