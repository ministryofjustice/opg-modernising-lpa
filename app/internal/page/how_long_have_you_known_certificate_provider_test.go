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

func TestGetHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	certificateProvider := CertificateProvider{RelationshipLength: "gte-2-years"}

	dataStore := &mockDataStore{data: Lpa{CertificateProvider: certificateProvider}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			HowLong:             "gte-2-years",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestPostHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			CertificateProvider: CertificateProvider{RelationshipLength: "gte-2-years"},
			Tasks:               Tasks{CertificateProvider: TaskCompleted},
		}).
		Return(nil)

	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howLongHaveYouKnownCertificateProviderData{
			App: appData,
			Errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadHowLongHaveYouKnownCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how-long": {"gte-2-years"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readHowLongHaveYouKnownCertificateProviderForm(r)

	assert.Equal(t, "gte-2-years", result.HowLong)
}

func TestHowLongHaveYouKnownCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howLongHaveYouKnownCertificateProviderForm
		errors map[string]string
	}{
		"gte-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "gte-2-years",
			},
			errors: map[string]string{},
		},
		"lt-2-years": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "lt-2-years",
			},
			errors: map[string]string{
				"how-long": "mustHaveKnownCertificateProviderTwoYears",
			},
		},
		"missing": {
			form: &howLongHaveYouKnownCertificateProviderForm{},
			errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		},
		"invalid": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				HowLong: "what",
			},
			errors: map[string]string{
				"how-long": "selectHowLongHaveYouKnownCertificateProvider",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
