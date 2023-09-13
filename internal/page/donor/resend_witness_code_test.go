package donor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetResendWitnessCode(t *testing.T) {
	for _, actorType := range []actor.Type{actor.TypeIndependentWitness, actor.TypeCertificateProvider} {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &resendWitnessCodeData{
					App: testAppData,
				}).
				Return(nil)

			err := ResendWitnessCode(template.Execute, &mockWitnessCodeSender{}, time.Now, actorType)(testAppData, w, r, &page.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetResendWitnessCodeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(template.Execute, &mockWitnessCodeSender{}, time.Now, actor.TypeCertificateProvider)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostResendWitnessCode(t *testing.T) {
	testcases := map[actor.Type]struct {
		redirect page.LpaPath
		method   string
	}{
		actor.TypeIndependentWitness: {
			redirect: page.Paths.WitnessingAsIndependentWitness,
			method:   "SendToIndependentWitness",
		},
		actor.TypeCertificateProvider: {
			redirect: page.Paths.WitnessingAsCertificateProvider,
			method:   "SendToCertificateProvider",
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpa := &page.Lpa{
				ID:                    "lpa-id",
				DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.DrivingLicencePaper},
			}

			witnessCodeSender := newMockWitnessCodeSender(t)
			witnessCodeSender.
				On(tc.method, r.Context(), lpa, testAppData.Localizer).
				Return(nil)

			err := ResendWitnessCode(nil, witnessCodeSender, time.Now, actorType)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostResendWitnessCodeWhenSendErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{Donor: actor.Donor{FirstNames: "john"}}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("SendToCertificateProvider", r.Context(), lpa, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(nil, witnessCodeSender, time.Now, actor.TypeCertificateProvider)(testAppData, w, r, lpa)

	assert.Equal(t, expectedError, err)
}

func TestPostResendWitnessCodeWhenTooRecentlySent(t *testing.T) {
	testcases := map[actor.Type]*page.Lpa{
		actor.TypeIndependentWitness: {
			Donor:                   actor.Donor{FirstNames: "john"},
			IndependentWitnessCodes: page.WitnessCodes{{Created: time.Now()}},
		},
		actor.TypeCertificateProvider: {
			Donor:                    actor.Donor{FirstNames: "john"},
			CertificateProviderCodes: page.WitnessCodes{{Created: time.Now()}},
		},
	}

	for actorType, lpa := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &resendWitnessCodeData{
					App:    testAppData,
					Errors: validation.With("request", validation.CustomError{Label: "pleaseWaitOneMinute"}),
				}).
				Return(nil)

			err := ResendWitnessCode(template.Execute, &mockWitnessCodeSender{}, time.Now, actorType)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
