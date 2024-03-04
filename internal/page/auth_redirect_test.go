package page

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestAuthRedirect(t *testing.T) {
	testcases := map[string]struct {
		session  *sesh.OneLoginSession
		redirect string
	}{
		"login": {
			session: &sesh.OneLoginSession{
				State:    "my-state",
				Nonce:    "my-nonce",
				Locale:   "en",
				Redirect: Paths.LoginCallback.Format(),
			},
			redirect: Paths.LoginCallback.Format(),
		},
		"login with nested route": {
			session: &sesh.OneLoginSession{
				State:    "my-state",
				Nonce:    "my-nonce",
				Locale:   "en",
				Redirect: Paths.IdentityWithOneLoginCallback.Format("123"),
				LpaID:    "123",
			},
			redirect: Paths.IdentityWithOneLoginCallback.Format("123"),
		},
		"welsh": {
			session: &sesh.OneLoginSession{
				State:    "my-state",
				Nonce:    "my-nonce",
				Locale:   "cy",
				Redirect: Paths.IdentityWithOneLoginCallback.Format("123"),
				LpaID:    "123",
			},
			redirect: "/cy" + Paths.IdentityWithOneLoginCallback.Format("123"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				OneLogin(r).
				Return(tc.session, nil)

			AuthRedirect(nil, sessionStore)(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect+"?code=auth-code&state=my-state", resp.Header.Get("Location"))
		})
	}
}

func TestAuthRedirectSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Info("problem retrieving onelogin session", slog.Any("err", expectedError))

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(nil, expectedError)

	AuthRedirect(logger, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthRedirectStateIncorrect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hello", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		Info("state incorrect")

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Redirect: Paths.LoginCallback.Format()}, nil)

	AuthRedirect(logger, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
