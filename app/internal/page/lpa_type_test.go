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

func TestGetLpaType(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := LpaType(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetLpaTypeFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{data: Lpa{Type: "pfa"}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			App:  appData,
			Type: "pfa",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := LpaType(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetLpaTypeWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := LpaType(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetLpaTypeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := LpaType(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestPostLpaType(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{Type: "pfa"}).
		Return(nil)

	form := url.Values{
		"lpa-type": {"pfa"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := LpaType(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whoIsTheLpaForPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostLpaTypeWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{Type: "pfa"}).
		Return(expectedError)

	form := url.Values{
		"lpa-type": {"pfa"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := LpaType(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostLpaTypeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			App: appData,
			Errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := LpaType(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadLpaTypeForm(t *testing.T) {
	form := url.Values{
		"lpa-type": {"pfa"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readLpaTypeForm(r)

	assert.Equal(t, "pfa", result.LpaType)
}

func TestLpaTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lpaTypeForm
		errors map[string]string
	}{
		"pfa": {
			form: &lpaTypeForm{
				LpaType: "pfa",
			},
			errors: map[string]string{},
		},
		"hw": {
			form: &lpaTypeForm{
				LpaType: "hw",
			},
			errors: map[string]string{},
		},
		"both": {
			form: &lpaTypeForm{
				LpaType: "both",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &lpaTypeForm{},
			errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		},
		"invalid": {
			form: &lpaTypeForm{
				LpaType: "what",
			},
			errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
