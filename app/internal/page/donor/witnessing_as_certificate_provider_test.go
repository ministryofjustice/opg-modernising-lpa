package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App:  testAppData,
			Lpa:  &page.Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
		Return(&page.Lpa{}, expectedError)

	template := &mockTemplate{}

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
		Return(&page.Lpa{
			CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App:  testAppData,
			Lpa:  &page.Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCode: page.WitnessCode{Code: "1234", Created: now},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			WitnessCode:            page.WitnessCode{Code: "1234", Created: now},
			CPWitnessCodeValidated: true,
			Submitted:              now,
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, lpaStore, nil, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWitnessingAsCertificateProviderWhenIdentityConfirmed(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider:         actor.CertificateProvider{Email: "name@example.com"},
			CertificateProviderUserData: identity.UserData{OK: true},
			WitnessCode:                 page.WitnessCode{Code: "1234", Created: now},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider:         actor.CertificateProvider{Email: "name@example.com"},
			CertificateProviderUserData: identity.UserData{OK: true},
			WitnessCode:                 page.WitnessCode{Code: "1234", Created: now},
			CPWitnessCodeValidated:      true,
			Submitted:                   now,
		}).
		Return(nil)

	shareCodeSender := &mockShareCodeSender{}
	shareCodeSender.
		On("Send", r.Context(), notify.CertificateProviderCertifyEmail, testAppData, "name@example.com", false).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, lpaStore, shareCodeSender, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, shareCodeSender)
}

func TestPostWitnessingAsCertificateProviderWhenShareCodeSendErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider:         actor.CertificateProvider{Email: "name@example.com"},
			CertificateProviderUserData: identity.UserData{OK: true},
			WitnessCode:                 page.WitnessCode{Code: "1234", Created: now},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider:         actor.CertificateProvider{Email: "name@example.com"},
			CertificateProviderUserData: identity.UserData{OK: true},
			WitnessCode:                 page.WitnessCode{Code: "1234", Created: now},
			CPWitnessCodeValidated:      true,
			Submitted:                   now,
		}).
		Return(nil)

	shareCodeSender := &mockShareCodeSender{}
	shareCodeSender.
		On("Send", r.Context(), notify.CertificateProviderCertifyEmail, testAppData, "name@example.com", false).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, lpaStore, shareCodeSender, func() time.Time { return now })(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, shareCodeSender)
}

func TestPostWitnessingAsCertificateProviderCodeTooOld(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCode: page.WitnessCode{Code: "1234", Created: invalidCreated},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCode: page.WitnessCode{Code: "1234", Created: invalidCreated},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCode: page.WitnessCode{Code: "1234", Created: invalidCreated},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCode: page.WitnessCode{Code: "1234", Created: invalidCreated},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCode: page.WitnessCode{Code: "1234", Created: now},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCode: page.WitnessCode{Code: "1234", Created: now},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Func, lpaStore, nil, time.Now)(testAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

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
