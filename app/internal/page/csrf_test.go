package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestPostMakeHandleCsrfTokenValid(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostMakeHandleCsrfTokensNotEqual(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"321"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPostMakeHandleCsrfTokenCookieValueEmpty(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"not-token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPostMakeHandleCsrfTokenErrorWhenDecodingSession(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, ExpectedError)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetMakeHandleCsrfSessionSavedWhenNew(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{IsNew: true}, nil)
	sessionStore.
		On("Save", r, w, &sessions.Session{
			IsNew:  true,
			Values: map[interface{}]interface{}{"token": "123"},
			Options: &sessions.Options{
				MaxAge:   86400,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
		}).
		Return(nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
