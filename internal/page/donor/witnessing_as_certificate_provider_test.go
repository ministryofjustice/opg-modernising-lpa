package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Joan"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{},
			Form:  &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Execute, nil, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsCertificateProvider(t *testing.T) {
	testcases := map[string]struct {
		certificateProvider *actor.CertificateProviderProvidedDetails
		err                 error
	}{
		"not found": {
			err: dynamo.NotFoundError{},
		},
		"empty": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
		},
		"identity not confirmed": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{
					FirstNames: "Barry",
				},
			},
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

			donor := &actor.DonorProvidedDetails{
				LpaID:                            "lpa-id",
				DonorIdentityUserData:            identity.UserData{OK: true},
				CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
				CertificateProvider:              actor.CertificateProvider{FirstNames: "Fred"},
				WitnessedByCertificateProviderAt: now,
				SignedAt:                         now,
			}

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.
				On("SendAttorneys", r.Context(), testAppData, donor).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), donor).
				Return(nil)

			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
				SessionID: "session-id",
				LpaID:     "lpa-id",
			})

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", ctx).
				Return(tc.certificateProvider, tc.err)

			err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:                    "lpa-id",
				DonorIdentityUserData:    identity.UserData{OK: true},
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
				CertificateProvider:      actor.CertificateProvider{FirstNames: "Fred"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostWitnessingAsCertificateProviderWhenIdentityConfirmed(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donor := &actor.DonorProvidedDetails{
		LpaID:                            "lpa-id",
		DonorIdentityUserData:            identity.UserData{OK: true},
		CertificateProvider:              actor.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
		WitnessedByCertificateProviderAt: now,
		SignedAt:                         now,
	}
	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), donor).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", ctx).
		Return(&actor.CertificateProviderProvidedDetails{IdentityUserData: identity.UserData{OK: true}}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderProvideCertificatePromptEmail, testAppData, false, donor).
		Return(nil)
	shareCodeSender.
		On("SendAttorneys", r.Context(), testAppData, donor).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                    "lpa-id",
		DonorIdentityUserData:    identity.UserData{OK: true},
		CertificateProvider:      actor.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YouHaveSubmittedYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenShareCodeSendToAttorneysErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donor := &actor.DonorProvidedDetails{
		LpaID:                            "lpa-id",
		DonorIdentityUserData:            identity.UserData{OK: true},
		CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
		CertificateProvider:              actor.CertificateProvider{FirstNames: "Fred"},
		WitnessedByCertificateProviderAt: now,
		SignedAt:                         now,
	}

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendAttorneys", r.Context(), testAppData, donor).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), donor).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", ctx).
		Return(nil, dynamo.NotFoundError{})

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                    "lpa-id",
		DonorIdentityUserData:    identity.UserData{OK: true},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
		CertificateProvider:      actor.CertificateProvider{FirstNames: "Fred"},
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

	donor := &actor.DonorProvidedDetails{
		CertificateProvider:              actor.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes:         actor.WitnessCodes{{Code: "1234", Created: now}},
		WitnessedByCertificateProviderAt: now,
		SignedAt:                         now,
	}
	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), donor).
		Return(nil)

	ctx := page.ContextWithSessionData(
		r.Context(),
		&page.SessionData{
			SessionID: "session-id",
			LpaID:     "lpa-id",
		},
	)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", ctx).
		Return(&actor.CertificateProviderProvidedDetails{IdentityUserData: identity.UserData{OK: true}}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderProvideCertificatePromptEmail, testAppData, false, donor).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, donorStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		CertificateProvider:      actor.CertificateProvider{Email: "name@example.com"},
		CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
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
	donorStore.
		On("Put", r.Context(), mock.MatchedBy(func(donor *actor.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
	donorStore.
		On("Put", r.Context(), mock.MatchedBy(func(donor *actor.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
	donorStore.
		On("Put", r.Context(), mock.MatchedBy(func(donor *actor.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
	donorStore.
		On("Put", r.Context(), mock.MatchedBy(func(donor *actor.DonorProvidedDetails) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Donor: &actor.DonorProvidedDetails{
				CertificateProviderCodes: actor.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, donorStore, nil, time.Now, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		WitnessCodeLimiter:       actor.NewLimiter(time.Minute, 0, 10),
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
