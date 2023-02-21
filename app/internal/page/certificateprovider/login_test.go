package certificateprovider

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

func TestCertificateProviderLogin(t *testing.T) {
	testcases := map[string]struct {
		url      string
		identity bool
	}{
		"identity": {
			url:      "/?sessionId=session-id&lpaId=lpa-id&identity=1",
			identity: true,
		},
		"sign in": {
			url:      "/?sessionId=session-id&lpaId=lpa-id",
			identity: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			client := newMockOneLoginClient(t)
			client.
				On("AuthCodeURL", "i am random", "i am random", "cy", tc.identity).
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
				"one-login": &sesh.OneLoginSession{
					State:               "i am random",
					Nonce:               "i am random",
					Locale:              "cy",
					CertificateProvider: true,
					Identity:            tc.identity,
					SessionID:           "session-id",
					LpaID:               "lpa-id",
				},
			}

			sessionStore.
				On("Save", r, w, session).
				Return(nil)

			Login(nil, client, sessionStore, func(int) string { return "i am random" })(page.AppData{Lang: localize.Cy, Paths: page.Paths}, w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "http://auth", resp.Header.Get("Location"))
		})
	}
}

func TestCertificateProviderLoginDefaultLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?sessionId=session-id&lpaId=lpa-id&identity=1", nil)

	client := newMockOneLoginClient(t)
	client.
		On("AuthCodeURL", "i am random", "i am random", "en", true).
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
		"one-login": &sesh.OneLoginSession{
			State:               "i am random",
			Nonce:               "i am random",
			Locale:              "en",
			CertificateProvider: true,
			Identity:            true,
			SessionID:           "session-id",
			LpaID:               "lpa-id",
		},
	}

	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	Login(nil, client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestCertificateProviderLoginWhenStoreSaveError(t *testing.T) {
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
