package page

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
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
				Redirect: PathLoginCallback.Format(),
			},
			redirect: PathLoginCallback.Format(),
		},
		"login with nested route": {
			session: &sesh.OneLoginSession{
				State:    "my-state",
				Nonce:    "my-nonce",
				Locale:   "en",
				Redirect: donor.PathIdentityWithOneLoginCallback.Format("123"),
				LpaID:    "123",
			},
			redirect: donor.PathIdentityWithOneLoginCallback.Format("123"),
		},
		"welsh": {
			session: &sesh.OneLoginSession{
				State:    "my-state",
				Nonce:    "my-nonce",
				Locale:   "cy",
				Redirect: donor.PathIdentityWithOneLoginCallback.Format("123"),
				LpaID:    "123",
			},
			redirect: "/cy" + donor.PathIdentityWithOneLoginCallback.Format("123"),
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
		InfoContext(r.Context(), "problem retrieving onelogin session", slog.Any("err", expectedError))

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
		InfoContext(r.Context(), "state incorrect")

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Redirect: PathLoginCallback.Format()}, nil)

	AuthRedirect(logger, sessionStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
