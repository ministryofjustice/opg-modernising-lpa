package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWitnessingAsCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App:   testAppData,
			Donor: &donordata.DonorProvidedDetails{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.DonorProvidedDetails{
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProvider: donordata.CertificateProvider{FirstNames: "Joan"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App:   testAppData,
			Donor: &donordata.DonorProvidedDetails{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProvider(t *testing.T) {
	testcases := map[string]struct {
		registeringWithCOP                bool
		expectedIdentityAndSignTaskStatus actor.IdentityTask
	}{
		"registering with COP": {
			registeringWithCOP:                true,
			expectedIdentityAndSignTaskStatus: actor.IdentityTaskPending,
		},
		"proved ID": {
			expectedIdentityAndSignTaskStatus: actor.IdentityTaskCompleted,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"witness-code": {"1234"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)
			now := time.Now()

			donor := &donordata.DonorProvidedDetails{
				LpaID:                            "lpa-id",
				DonorIdentityUserData:            identity.UserData{Status: identity.StatusConfirmed},
				CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
				CertificateProvider:              donordata.CertificateProvider{FirstNames: "Fred"},
				WitnessedByCertificateProviderAt: now,
				Tasks: donordata.DonorTasks{
					ConfirmYourIdentityAndSign: tc.expectedIdentityAndSignTaskStatus,
					PayForLpa:                  actor.PaymentTaskCompleted,
				},
				RegisteringWithCourtOfProtection: tc.registeringWithCOP,
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), donor).
				Return(nil)

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.EXPECT().
				SendCertificateProviderPrompt(r.Context(), testAppData, donor).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendLpa(r.Context(), donor).
				Return(nil)

			err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:                            "lpa-id",
				DonorIdentityUserData:            identity.UserData{Status: identity.StatusConfirmed},
				CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
				CertificateProvider:              donordata.CertificateProvider{FirstNames: "Fred"},
				Tasks:                            donordata.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				RegisteringWithCourtOfProtection: tc.registeringWithCOP,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}

}

func TestPostWitnessingAsCertificateProviderWhenPaymentPending(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donor := &donordata.DonorProvidedDetails{
		LpaID:                            "lpa-id",
		DonorIdentityUserData:            identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:              donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
		WitnessedByCertificateProviderAt: now,
		Tasks: donordata.DonorTasks{
			PayForLpa:                  actor.PaymentTaskPending,
			ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
		},
	}
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), donor).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, donorStore, nil, nil, func() time.Time { return now })(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:                    "lpa-id",
		DonorIdentityUserData:    identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:      donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
		Tasks:                    donordata.DonorTasks{PayForLpa: actor.PaymentTaskPending},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenSendLpaErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(r.Context(), mock.Anything).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:                    "lpa-id",
		DonorIdentityUserData:    identity.UserData{Status: identity.StatusConfirmed},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
		CertificateProvider:      donordata.CertificateProvider{FirstNames: "Fred"},
		Tasks:                    donordata.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
	})
	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingAsCertificateProviderWhenShareCodeSendToCertificateProviderErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), testAppData, mock.Anything).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, nil, func() time.Time { return now })(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProvider:      donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
		Tasks:                    donordata.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
	})

	assert.Equal(t, expectedError, err)
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

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProviderCodeDoesNotMatch(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProviderWhenCodeExpired(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProviderCodeLimitBreached(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, time.Now)(testAppData, w, r, &donordata.DonorProvidedDetails{
		WitnessCodeLimiter:       donordata.NewLimiter(time.Minute, 0, 10),
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
