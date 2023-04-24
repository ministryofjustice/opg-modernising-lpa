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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App:  testAppData,
			Lpa:  &page.Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	template := newMockTemplate(t)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProviderDetails: page.CertificateProviderDetails{FirstNames: "Joan"},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				CertificateProviderDetails: page.CertificateProviderDetails{FirstNames: "Joan"},
			},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App:  testAppData,
			Lpa:  &page.Lpa{},
			Form: &witnessingAsCertificateProviderForm{},
		}).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
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
	now := time.Now()

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin},
			WitnessCodes:          page.WitnessCodes{{Code: "1234", Created: now}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			DonorIdentityUserData:  identity.UserData{OK: true, Provider: identity.OneLogin},
			WitnessCodes:           page.WitnessCodes{{Code: "1234", Created: now}},
			CPWitnessCodeValidated: true,
			Submitted:              now,
		}).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, nil)

	err := WitnessingAsCertificateProvider(nil, lpaStore, nil, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenIdentityConfirmed(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	lpa := &page.Lpa{
		DonorIdentityUserData:      identity.UserData{OK: true, Provider: identity.OneLogin},
		CertificateProviderDetails: page.CertificateProviderDetails{Email: "name@example.com"},
		WitnessCodes:               page.WitnessCodes{{Code: "1234", Created: now}},
		CPWitnessCodeValidated:     true,
		Submitted:                  now,
	}
	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			DonorIdentityUserData:      identity.UserData{OK: true, Provider: identity.OneLogin},
			CertificateProviderDetails: page.CertificateProviderDetails{Email: "name@example.com"},
			WitnessCodes:               page.WitnessCodes{{Code: "1234", Created: now}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), lpa).
		Return(nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
	})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{IdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin}}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderReturnEmail, testAppData, false, lpa).
		Return(nil)

	err := WitnessingAsCertificateProvider(nil, lpaStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YouHaveSubmittedYourLpa, resp.Header.Get("Location"))
}

func TestPostWitnessingAsCertificateProviderWhenShareCodeSendErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	lpa := &page.Lpa{
		CertificateProviderDetails: page.CertificateProviderDetails{Email: "name@example.com"},
		WitnessCodes:               page.WitnessCodes{{Code: "1234", Created: now}},
		CPWitnessCodeValidated:     true,
		Submitted:                  now,
	}
	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProviderDetails: page.CertificateProviderDetails{Email: "name@example.com"},
			WitnessCodes:               page.WitnessCodes{{Code: "1234", Created: now}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), lpa).
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
		On("Get", ctx).
		Return(&actor.CertificateProvider{IdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin}}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderReturnEmail, testAppData, false, lpa).
		Return(expectedError)

	err := WitnessingAsCertificateProvider(nil, lpaStore, shareCodeSender, func() time.Time { return now }, certificateProviderStore)(testAppData, w, r)

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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.MatchedBy(func(lpa *page.Lpa) bool {
			lpa.WitnessCodeLimiter = nil
			return assert.Equal(t, lpa, &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCodes: page.WitnessCodes{{Code: "1234", Created: now}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.MatchedBy(func(lpa *page.Lpa) bool {
			lpa.WitnessCodeLimiter = nil
			return assert.Equal(t, lpa, &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.MatchedBy(func(lpa *page.Lpa) bool {
			lpa.WitnessCodeLimiter = nil
			return assert.Equal(t, lpa, &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			WitnessCodeLimiter: page.NewLimiter(time.Minute, 0, 10),
			WitnessCodes:       page.WitnessCodes{{Code: "1234", Created: now}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), mock.MatchedBy(func(lpa *page.Lpa) bool {
			lpa.WitnessCodeLimiter = nil
			return assert.Equal(t, lpa, &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingAsCertificateProviderData{
			App: testAppData,
			Lpa: &page.Lpa{
				WitnessCodes: page.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"}),
			Form:   &witnessingAsCertificateProviderForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsCertificateProvider(template.Execute, lpaStore, nil, time.Now, nil)(testAppData, w, r)
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
