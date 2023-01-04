package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWitnessingAsCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	template := &mockTemplate{}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			CertificateProvider: CertificateProvider{FirstNames: "Joan"},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{
				CertificateProvider: CertificateProvider{FirstNames: "Joan"},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingAsCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			WitnessCode: WitnessCode{Code: "1234", Created: now},
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			WitnessCode:            WitnessCode{Code: "1234", Created: now},
			CPWitnessCodeValidated: true,
		}).
		Return(nil)

	form := url.Values{
		"witness-code": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WitnessingAsCertificateProvider(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWitnessingAsCertificateProviderCodeTooOld(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			WitnessCode: WitnessCode{Code: "1234", Created: invalidCreated},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{
				WitnessCode: WitnessCode{Code: "1234", Created: invalidCreated},
			},
			Errors: map[string]string{
				"witness-code": "witnessCodeExpired",
			},
			Form: &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	form := url.Values{
		"witness-code": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingAsCertificateProviderCodeDoesNotMatch(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			WitnessCode: WitnessCode{Code: "1234", Created: now},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{
				WitnessCode: WitnessCode{Code: "1234", Created: now},
			},
			Errors: map[string]string{
				"witness-code": "witnessCodeDoesNotMatch",
			},
			Form: &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	form := url.Values{
		"witness-code": {"4321"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestReadWitnessingAsCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWitnessingAsCertificateProviderForm(r)

	assert.Equal(t, "1234", result.Code)
}

func TestWitnessingAsCertificateProviderValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *witnessingAsCertificateProviderForm
		errors map[string]string
	}{
		"valid numeric": {
			form: &witnessingAsCertificateProviderForm{
				Code: "1234",
			},
			errors: map[string]string{},
		},
		"valid alpha": {
			form: &witnessingAsCertificateProviderForm{
				Code: "aBcD",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &witnessingAsCertificateProviderForm{},
			errors: map[string]string{
				"witness-code": "enterWitnessCode",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
