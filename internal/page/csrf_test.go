package page

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

func TestPostValidateCsrf(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostValidateCsrfWhenMultipartForm(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file, _ := os.Open("testdata/dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "whatever.pdf")
	io.Copy(part, file)

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", &buf)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"token": "123"}}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostValidateCsrfInvalid(t *testing.T) {
	testcases := map[string]struct {
		csrf   string
		cookie string
	}{
		"not equal": {
			csrf:   "321",
			cookie: "token",
		},
		"cookie missing": {
			csrf:   "123",
			cookie: "not-token",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			form := url.Values{
				"csrf": {tc.csrf},
			}
			r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(r, "csrf").
				Return(&sessions.Session{Values: map[interface{}]interface{}{tc.cookie: "123"}}, nil)

			errorHandler := newMockErrorHandler(t)
			errorHandler.EXPECT().
				Execute(w, r, ErrCsrfInvalid).
				Return()

			ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, errorHandler.Execute).ServeHTTP(w, r)
		})
	}
}

func TestPostValidateCsrfWhenInvalidMultipartForm(t *testing.T) {
	testcases := map[string]struct {
		fieldName    string
		fieldContent string
	}{
		"bad value": {
			fieldName:    "csrf",
			fieldContent: "hey",
		},
		"wrong field": {
			fieldName:    "not-csrf",
			fieldContent: "123456789012",
		},
		"over size value": {
			fieldName:    "csrf",
			fieldContent: "1234567890123",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, _ := writer.CreateFormField(tc.fieldName)
			io.WriteString(part, tc.fieldContent)

			writer.Close()

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Type", writer.FormDataContentType())

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(r, "csrf").
				Return(&sessions.Session{Values: map[any]any{"token": "123456789012"}}, nil)

			errorHandler := newMockErrorHandler(t)
			errorHandler.EXPECT().
				Execute(w, r, ErrCsrfInvalid).
				Return()

			ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, errorHandler.Execute).ServeHTTP(w, r)
		})
	}
}

func TestPostValidateCsrfErrorWhenDecodingSession(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {"123"},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "csrf").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError).
		Return()

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, errorHandler.Execute).ServeHTTP(w, r)
}

func TestGetValidateCsrfSessionSavedWhenNew(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "csrf").
		Return(&sessions.Session{IsNew: true}, nil)
	sessionStore.EXPECT().
		Save(r, w, &sessions.Session{
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

	ValidateCsrf(http.NotFoundHandler(), sessionStore, MockRandom, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
