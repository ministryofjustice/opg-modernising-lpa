package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.
		On("AuthCodeURL", "i am random", "i am random", "cy", false).
		Return("http://auth")

	sessionStore := newMockSessionStore(t)

	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity: true,
			LpaID:    "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	loginSession := sessions.NewSession(sessionStore, "params")

	loginSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	loginSession.Values = map[any]any{
		"one-login": &sesh.OneLoginSession{
			State:    "i am random",
			Nonce:    "i am random",
			Locale:   "cy",
			LpaID:    "lpa-id",
			Redirect: page.Paths.CertificateProvider.LoginCallback,
		},
	}

	sessionStore.
		On("Save", r, w, loginSession).
		Return(nil)

	err := Login(nil, client, sessionStore, func(int) string { return "i am random" })(page.AppData{Lang: localize.Cy, Paths: page.Paths}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestCertificateProviderLoginDefaultLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.
		On("AuthCodeURL", "i am random", "i am random", "en", false).
		Return("http://auth")

	sessionStore := newMockSessionStore(t)

	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity: true,
			LpaID:    "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	loginSession := sessions.NewSession(sessionStore, "params")

	loginSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	loginSession.Values = map[any]any{
		"one-login": &sesh.OneLoginSession{
			State:    "i am random",
			Nonce:    "i am random",
			Locale:   "en",
			LpaID:    "lpa-id",
			Redirect: page.Paths.CertificateProvider.LoginCallback,
		},
	}

	sessionStore.
		On("Save", r, w, loginSession).
		Return(nil)

	err := Login(nil, client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
}

func TestCertificateProviderLoginWhenStoreGetError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "shareCode").
		Return(sessions.NewSession(sessionStore, "shareCode"), expectedError)

	err := Login(logger, nil, sessionStore, func(int) string { return "i am random" })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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

	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity: true,
			LpaID:    "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	err := Login(logger, client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
