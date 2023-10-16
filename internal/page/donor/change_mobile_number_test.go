package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChangeMobileNumber(t *testing.T) {
	for actorType, firstNames := range map[actor.Type]string{
		actor.TypeIndependentWitness:  "Independent",
		actor.TypeCertificateProvider: "Certificate",
	} {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &changeMobileNumberData{
					App:        testAppData,
					Form:       &changeMobileNumberForm{},
					ActorType:  actorType,
					FirstNames: firstNames,
				}).
				Return(nil)

			err := ChangeMobileNumber(template.Execute, newMockWitnessCodeSender(t), actorType)(testAppData, w, r, &page.Lpa{
				CertificateProvider: actor.CertificateProvider{FirstNames: "Certificate", LastName: "Provided"},
				IndependentWitness:  actor.IndependentWitness{FirstNames: "Independent", LastName: "Witness"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChangeMobileNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		lpa       *page.Lpa
		actorType actor.Type
	}{
		"certificate provider uk mobile": {
			lpa: &page.Lpa{
				CertificateProvider: actor.CertificateProvider{
					Mobile: "07777",
				},
			},
			actorType: actor.TypeCertificateProvider,
		},
		"certificate provider non-uk mobile": {
			lpa: &page.Lpa{
				CertificateProvider: actor.CertificateProvider{
					Mobile:         "07777",
					HasNonUKMobile: true,
				},
			},
			actorType: actor.TypeCertificateProvider,
		},
		"independent witness uk mobile": {
			lpa: &page.Lpa{
				IndependentWitness: actor.IndependentWitness{
					Mobile: "07777",
				},
			},
			actorType: actor.TypeIndependentWitness,
		},
		"independent witness non-uk mobile": {
			lpa: &page.Lpa{
				IndependentWitness: actor.IndependentWitness{
					Mobile:         "07777",
					HasNonUKMobile: true,
				},
			},
			actorType: actor.TypeIndependentWitness,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &changeMobileNumberData{
					App:       testAppData,
					Form:      &changeMobileNumberForm{},
					ActorType: tc.actorType,
				}).
				Return(nil)

			err := ChangeMobileNumber(template.Execute, newMockWitnessCodeSender(t), tc.actorType)(testAppData, w, r, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChangeMobileNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ChangeMobileNumber(template.Execute, newMockWitnessCodeSender(t), actor.TypeCertificateProvider)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChangeMobileNumber(t *testing.T) {
	testCases := map[string]struct {
		form      url.Values
		actorType actor.Type
		lpa       *page.Lpa
		send      string
		redirect  page.LpaPath
	}{
		"certificate provider valid": {
			form: url.Values{
				"mobile": {"07535111111"},
			},
			actorType: actor.TypeCertificateProvider,
			lpa: &page.Lpa{
				ID: "lpa-id",
				CertificateProvider: actor.CertificateProvider{
					Mobile: "07535111111",
				},
				DonorIdentityUserData: identity.UserData{
					OK:       true,
					Provider: identity.DrivingLicencePaper,
				},
			},
			send:     "SendToCertificateProvider",
			redirect: page.Paths.WitnessingAsCertificateProvider,
		},
		"certificate provider valid non uk mobile": {
			form: url.Values{
				"has-non-uk-mobile": {"1"},
				"non-uk-mobile":     {"+337575757"},
			},
			actorType: actor.TypeCertificateProvider,
			lpa: &page.Lpa{
				ID: "lpa-id",
				CertificateProvider: actor.CertificateProvider{
					Mobile:         "+337575757",
					HasNonUKMobile: true,
				},
				DonorIdentityUserData: identity.UserData{
					OK:       true,
					Provider: identity.DrivingLicencePaper,
				},
			},
			send:     "SendToCertificateProvider",
			redirect: page.Paths.WitnessingAsCertificateProvider,
		},
		"independent witness valid": {
			form: url.Values{
				"mobile": {"07535111111"},
			},
			actorType: actor.TypeIndependentWitness,
			lpa: &page.Lpa{
				ID: "lpa-id",
				IndependentWitness: actor.IndependentWitness{
					Mobile: "07535111111",
				},
				DonorIdentityUserData: identity.UserData{
					OK:       true,
					Provider: identity.DrivingLicencePaper,
				},
			},
			send:     "SendToIndependentWitness",
			redirect: page.Paths.WitnessingAsIndependentWitness,
		},
		"independent witness valid non uk mobile": {
			form: url.Values{
				"has-non-uk-mobile": {"1"},
				"non-uk-mobile":     {"+337575757"},
			},
			actorType: actor.TypeIndependentWitness,
			lpa: &page.Lpa{
				ID: "lpa-id",
				IndependentWitness: actor.IndependentWitness{
					Mobile:         "+337575757",
					HasNonUKMobile: true,
				},
				DonorIdentityUserData: identity.UserData{
					OK:       true,
					Provider: identity.DrivingLicencePaper,
				},
			},
			send:     "SendToIndependentWitness",
			redirect: page.Paths.WitnessingAsIndependentWitness,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			witnessCodeSender := newMockWitnessCodeSender(t)
			witnessCodeSender.
				On(tc.send, r.Context(), tc.lpa, testAppData.Localizer).
				Return(nil)

			err := ChangeMobileNumber(nil, witnessCodeSender, tc.actorType)(testAppData, w, r, &page.Lpa{
				ID: "lpa-id",
				DonorIdentityUserData: identity.UserData{
					OK:       true,
					Provider: identity.DrivingLicencePaper,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostChangeMobileNumberWhenSendErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChangeMobileNumber(nil, witnessCodeSender, actor.TypeCertificateProvider)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostChangeMobileNumberWhenSendErrorsWithTooManyRequests(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
		Return(page.ErrTooManyWitnessCodeRequests)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *changeMobileNumberData) bool {
			return assert.Equal(t, validation.With("request", validation.CustomError{Label: "pleaseWaitOneMinute"}), data.Errors)
		})).
		Return(nil)

	err := ChangeMobileNumber(template.Execute, witnessCodeSender, actor.TypeCertificateProvider)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChangeMobileNumberWhenValidationError(t *testing.T) {
	form := url.Values{
		"mobile": {"xyz"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *changeMobileNumberData) bool {
			return assert.Equal(t, validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}), data.Errors)
		})).
		Return(nil)

	err := ChangeMobileNumber(template.Execute, newMockWitnessCodeSender(t), actor.TypeCertificateProvider)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadChangeMobileNumberForm(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChangeMobileNumberForm(r)

	assert.Equal(t, "07535111111", result.Mobile)
}

func TestChangeMobileNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *changeMobileNumberForm
		errors validation.List
	}{
		"valid": {
			form: &changeMobileNumberForm{
				Mobile: "07535111111",
			},
		},
		"missing all": {
			form: &changeMobileNumberForm{},
			errors: validation.
				With("mobile", validation.EnterError{Label: "aUKMobileNumber"}),
		},
		"missing when non uk mobile": {
			form: &changeMobileNumberForm{HasNonUKMobile: true},
			errors: validation.
				With("non-uk-mobile", validation.EnterError{Label: "aMobilePhoneNumber"}),
		},
		"invalid incorrect mobile format": {
			form: &changeMobileNumberForm{
				Mobile: "0753511111",
			},
			errors: validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
		"invalid non uk mobile format": {
			form: &changeMobileNumberForm{
				HasNonUKMobile: true,
				NonUKMobile:    "0753511111",
			},
			errors: validation.With("non-uk-mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
