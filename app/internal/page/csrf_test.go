package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostMakeHandleCsrfTokenValid(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionsStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestPostMakeHandleCsrfTokensNotEqual(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"321"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionsStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestPostMakeHandleCsrfTokenCookieValueEmpty(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"not-token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionsStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestPostMakeHandleCsrfTokenErrorWhenDecodingSession(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, ExpectedError)

	ValidateCsrf(http.NotFoundHandler(), sessionsStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestGetMakeHandleCsrfSessionSavedWhenNew(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{IsNew: true}, nil)
	sessionsStore.
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

	ValidateCsrf(http.NotFoundHandler(), sessionsStore, MockRandom).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}
