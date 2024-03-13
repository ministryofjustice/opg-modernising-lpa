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

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestPostValidateCsrf(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {testRandomString},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Csrf(r).
		Return(&sesh.CsrfSession{Token: testRandomString}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostValidateCsrfWhenMultipartForm(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, testRandomString)

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
		Csrf(r).
		Return(&sesh.CsrfSession{Token: testRandomString}, nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostValidateCsrfInvalid(t *testing.T) {
	testcases := map[string]*sesh.CsrfSession{
		"not equal": {
			Token: "321",
		},
		"cookie missing": {},
	}

	for name, session := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			form := url.Values{
				"csrf": {testRandomString},
			}
			r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Csrf(r).
				Return(session, nil)

			errorHandler := newMockErrorHandler(t)
			errorHandler.EXPECT().
				Execute(w, r, ErrCsrfInvalid).
				Return()

			ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, errorHandler.Execute).ServeHTTP(w, r)
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
				Csrf(r).
				Return(&sesh.CsrfSession{Token: "123456789012"}, nil)

			errorHandler := newMockErrorHandler(t)
			errorHandler.EXPECT().
				Execute(w, r, ErrCsrfInvalid).
				Return()

			ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, errorHandler.Execute).ServeHTTP(w, r)
		})
	}
}

func TestPostValidateCsrfErrorWhenDecodingSession(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"csrf": {testRandomString},
	}
	r, _ := http.NewRequest(http.MethodPost, "/path?a=b", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Csrf(r).
		Return(nil, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError).
		Return()

	ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, errorHandler.Execute).ServeHTTP(w, r)
}

func TestGetValidateCsrfSessionSavedWhenNew(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Csrf(r).
		Return(nil, expectedError)
	sessionStore.EXPECT().
		SetCsrf(r, w, &sesh.CsrfSession{Token: testRandomString}).
		Return(nil)

	ValidateCsrf(http.NotFoundHandler(), sessionStore, testRandomStringFn, nil).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
