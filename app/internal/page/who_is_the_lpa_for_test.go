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
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhoIsTheLpaFor(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetWhoIsTheLpaForWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhoIsTheLpaFor(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetWhoIsTheLpaForFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{data: Lpa{WhoFor: "me"}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App:    appData,
			WhoFor: "me",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhoIsTheLpaFor(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWhoIsTheLpaForWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhoIsTheLpaFor(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostWhoIsTheLpaFor(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{WhoFor: "me"}).
		Return(nil)

	form := url.Values{
		"who-for": {"me"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhoIsTheLpaFor(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, lpaTypePath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostWhoIsTheLpaForWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{WhoFor: "me"}).
		Return(expectedError)

	form := url.Values{
		"who-for": {"me"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhoIsTheLpaFor(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostWhoIsTheLpaForWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
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

	err := WhoIsTheLpaFor(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
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
