package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWitnessingAsCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App:  appData,
			Lpa:  &Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	template := &mockTemplate{}

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: appData,
			Lpa: &Lpa{
				CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingAsCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App:  appData,
			Lpa:  &Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingAsCertificateProvider(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			WitnessCode: WitnessCode{Code: "1234", Created: now},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			WitnessCode:            WitnessCode{Code: "1234", Created: now},
			CPWitnessCodeValidated: true,
			Submitted:              now,
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, lpaStore, func() time.Time { return now })(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWitnessingAsCertificateProviderCodeTooOld(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingAsCertificateProviderExpiryTrumpsMismatch(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingAsCertificateProviderCodeDoesNotMatch(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
			Errors: validation.With("witness-code", validation.CustomError{"witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, time.Now)(appData, w, r)
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
		errors validation.List
	}{
		"valid numeric": {
			form: &witnessingAsCertificateProviderForm{
				Code: "1234",
			},
		},
		"valid alpha": {
			form: &witnessingAsCertificateProviderForm{
				Code: "aBcD",
			},
		},
		"missing": {
			form:   &witnessingAsCertificateProviderForm{},
			errors: validation.With("witness-code", validation.EnterError{Label: "theCodeWeSentCertificateProvider"}),
		},
		"too long": {
			form: &witnessingAsCertificateProviderForm{
				Code: "12345",
			},
			errors: validation.With("witness-code", validation.StringLengthError{Label: "theCodeWeSentCertificateProvider", Length: 4}),
		},
		"too short": {
			form: &witnessingAsCertificateProviderForm{
				Code: "123",
			},
			errors: validation.With("witness-code", validation.StringLengthError{Label: "theCodeWeSentCertificateProvider", Length: 4}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
