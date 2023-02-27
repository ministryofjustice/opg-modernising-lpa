package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	testcases := map[localize.Lang]string{
		localize.En: "en",
		localize.Cy: "cy",
	}

	for lang, str := range testcases {
		t.Run(str, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			client := newMockOneLoginClient(t)
			client.
				On("AuthCodeURL", "i am random", "i am random", str, false).
				Return("http://auth")

			sessionStore := newMockSessionStore(t)

			session := sessions.NewSession(sessionStore, "params")

			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   600,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   true,
			}
			session.Values = map[any]any{
				"one-login": &sesh.OneLoginSession{State: "i am random", Nonce: "i am random", Locale: str},
			}

			sessionStore.
				On("Save", r, w, session).
				Return(nil)

			Login(nil, client, sessionStore, func(int) string { return "i am random" })(page.AppData{Lang: lang}, w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "http://auth", resp.Header.Get("Location"))
		})
	}
}

func TestLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	client := newMockOneLoginClient(t)
	client.
		On("AuthCodeURL", "i am random", "i am random", "en", false).
		Return("http://auth?locale=en")

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	Login(logger, client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
