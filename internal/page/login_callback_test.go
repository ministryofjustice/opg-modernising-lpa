package page

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
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
		"Sub exists":         {subExists: true, expectedRedirect: PathDashboard},
		"Sub does not exist": {subExists: false, expectedRedirect: PathAttorneyEnterReferenceNumber},
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

			session := &sesh.LoginSession{
				IDToken: "id-token",
				Sub:     "random",
				Email:   "name@example.com",
			}

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
				SetLogin(r, w, session).
				Return(nil)

			dashboardStore := newMockDashboardStore(t)
			dashboardStore.EXPECT().
				SubExistsForActorType(r.Context(), base64.StdEncoding.EncodeToString([]byte("random")), actor.TypeAttorney).
				Return(tc.subExists, nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				InfoContext(r.Context(), "login", slog.String("session_id", session.SessionID()))

			err := LoginCallback(logger, client, sessionStore, PathAttorneyEnterReferenceNumber, dashboardStore, actor.TypeAttorney)(appcontext.Data{}, w, r)
			assert.Nil(t, err)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestLoginCallbackWhenErrorReturned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?error=hey&error_description=this%20is%20why&state=my-state", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "login error", slog.String("error", "hey"), slog.String("error_description", "this is why"))

	err := LoginCallback(logger, nil, nil, PathAttorneyEnterReferenceNumber, nil, actor.TypeAttorney)(appcontext.Data{}, w, r)
	assert.Equal(t, errors.New("access denied"), err)
}

func TestLoginCallbackSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(nil, expectedError)

	err := LoginCallback(nil, nil, sessionStore, PathAttorneyLoginCallback, nil, actor.TypeAttorney)(appcontext.Data{}, w, r)
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
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: PathLoginCallback.Format()}, nil)

	err := LoginCallback(nil, client, sessionStore, PathLoginCallback, nil, actor.TypeAttorney)(appcontext.Data{}, w, r)
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
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: PathLoginCallback.Format()}, nil)

	err := LoginCallback(nil, client, sessionStore, PathLoginCallback, nil, actor.TypeAttorney)(appcontext.Data{}, w, r)
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
			Redirect: PathLoginCallback.Format(),
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, PathLoginCallback, nil, actor.TypeAttorney)(appcontext.Data{}, w, r)
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
			Redirect: PathLoginCallback.Format(),
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		SubExistsForActorType(r.Context(), mock.Anything, mock.Anything).
		Return(false, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, PathLoginCallback, dashboardStore, actor.TypeAttorney)(appcontext.Data{}, w, r)
	assert.Equal(t, expectedError, err)
}
