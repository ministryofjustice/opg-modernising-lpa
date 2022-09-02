package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhoIsTheLpaFor(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	WhoIsTheLpaFor(nil, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWhoIsTheLpaForWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	WhoIsTheLpaFor(logger, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}

func TestPostWhoIsTheLpaFor(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{WhoFor: "me"}).
		Return(nil)

	form := url.Values{
		"who-for": {"me"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	WhoIsTheLpaFor(nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donorDetailsPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostWhoIsTheLpaForWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: appData,
			Errors: map[string]string{
				"who-for": "selectWhoFor",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	WhoIsTheLpaFor(nil, template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWhoIsTheLpaForForm(t *testing.T) {
	form := url.Values{
		"who-for": {"me"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWhoIsTheLpaForForm(r)

	assert.Equal(t, "me", result.WhoFor)
}

func TestWhoIsTheLpaForFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whoIsTheLpaForForm
		errors map[string]string
	}{
		"me": {
			form: &whoIsTheLpaForForm{
				WhoFor: "me",
			},
			errors: map[string]string{},
		},
		"someone-else": {
			form: &whoIsTheLpaForForm{
				WhoFor: "someone-else",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &whoIsTheLpaForForm{},
			errors: map[string]string{
				"who-for": "selectWhoFor",
			},
		},
		"invalid": {
			form: &whoIsTheLpaForForm{
				WhoFor: "what",
			},
			errors: map[string]string{
				"who-for": "selectWhoFor",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
