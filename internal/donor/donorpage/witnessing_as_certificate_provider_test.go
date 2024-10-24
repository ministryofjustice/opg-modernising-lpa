package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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
			Donor: &donordata.Provided{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.Provided{})
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
			Donor: &donordata.Provided{
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.Provided{
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
			Donor: &donordata.Provided{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, nil, nil, time.Now)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProvider(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:                            "lpa-id",
		LpaUID:                           "lpa-uid",
		IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
		CertificateProviderCodes:         donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		CertificateProvider:              donordata.CertificateProvider{FirstNames: "Fred"},
		WitnessedByCertificateProviderAt: testNow,
		Tasks: donordata.Tasks{
			PayForLpa:  task.PaymentStateCompleted,
			SignTheLpa: task.StateCompleted,
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), provided).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), testAppData, provided).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(r.Context(), event.CertificateProviderStarted{
			UID: "lpa-uid",
		}).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(r.Context(), provided).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, lpaStoreClient, eventClient, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:                    "lpa-id",
		LpaUID:                   "lpa-uid",
		IdentityUserData:         identity.UserData{Status: identity.StatusConfirmed},
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		CertificateProvider:      donordata.CertificateProvider{FirstNames: "Fred"},
		Tasks:                    donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
		WitnessCodeLimiter:       testLimiter(),
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathYouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenPaymentPending(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:                            "lpa-id",
		IdentityUserData:                 identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:              donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes:         donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		WitnessedByCertificateProviderAt: testNow,
		Tasks: donordata.Tasks{
			PayForLpa:  task.PaymentStatePending,
			SignTheLpa: task.StateCompleted,
		},
	}
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), provided).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, donorStore, nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:                    "lpa-id",
		IdentityUserData:         identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:      donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		Tasks:                    donordata.Tasks{PayForLpa: task.PaymentStatePending},
		WitnessCodeLimiter:       testLimiter(),
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathYouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenSendLpaErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(r.Context(), mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(r.Context(), mock.Anything).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, lpaStoreClient, eventClient, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:                    "lpa-id",
		IdentityUserData:         identity.UserData{Status: identity.StatusConfirmed},
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		CertificateProvider:      donordata.CertificateProvider{FirstNames: "Fred"},
		Tasks:                    donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
		WitnessCodeLimiter:       testLimiter(),
	})
	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingAsCertificateProviderWhenEventClientErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(r.Context(), mock.Anything).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, nil, eventClient, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:                    "lpa-id",
		IdentityUserData:         identity.UserData{Status: identity.StatusConfirmed},
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		CertificateProvider:      donordata.CertificateProvider{FirstNames: "Fred"},
		Tasks:                    donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
		WitnessCodeLimiter:       testLimiter(),
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

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), testAppData, mock.Anything).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		CertificateProvider:      donordata.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		Tasks:                    donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
		WitnessCodeLimiter:       testLimiter(),
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

	invalidCreated := testNow.Add(-45 * time.Minute)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
		WitnessCodeLimiter:       testLimiter(),
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

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
		WitnessCodeLimiter:       testLimiter(),
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

	invalidCreated := testNow.Add(-45 * time.Minute)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
		WitnessCodeLimiter:       testLimiter(),
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

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &donordata.Provided{
				CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		WitnessCodeLimiter:       donordata.NewLimiter(time.Minute, 0, 10),
		CertificateProviderCodes: donordata.WitnessCodes{{Code: "1234", Created: testNow}},
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
