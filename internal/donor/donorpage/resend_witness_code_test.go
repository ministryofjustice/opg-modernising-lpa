package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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
			template.EXPECT().
				Execute(w, &resendWitnessCodeData{
					App: testAppData,
				}).
				Return(nil)

			err := ResendWitnessCode(template.Execute, &mockWitnessCodeSender{}, actorType)(testAppData, w, r, &donordata.Provided{})
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
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(template.Execute, &mockWitnessCodeSender{}, actor.TypeCertificateProvider)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostResendWitnessCode(t *testing.T) {
	testcases := map[actor.Type]struct {
		redirect donor.Path
		method   string
	}{
		actor.TypeIndependentWitness: {
			redirect: donor.PathWitnessingAsIndependentWitness,
			method:   "SendToIndependentWitness",
		},
		actor.TypeCertificateProvider: {
			redirect: donor.PathWitnessingAsCertificateProvider,
			method:   "SendToCertificateProvider",
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donor := &donordata.Provided{
				LpaID:                 "lpa-id",
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			}

			witnessCodeSender := newMockWitnessCodeSender(t)
			witnessCodeSender.
				On(tc.method, r.Context(), donor, testAppData.Localizer).
				Return(nil)

			err := ResendWitnessCode(nil, witnessCodeSender, actorType)(testAppData, w, r, donor)
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

	donor := &donordata.Provided{Donor: donordata.Donor{FirstNames: "john"}}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.EXPECT().
		SendToCertificateProvider(r.Context(), donor, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(nil, witnessCodeSender, actor.TypeCertificateProvider)(testAppData, w, r, donor)

	assert.Equal(t, expectedError, err)
}

func TestPostResendWitnessCodeWhenTooRecentlySent(t *testing.T) {
	testcases := map[actor.Type]struct {
		donor *donordata.Provided
		send  string
	}{
		actor.TypeIndependentWitness: {
			donor: &donordata.Provided{
				Donor:                   donordata.Donor{FirstNames: "john"},
				IndependentWitnessCodes: donordata.WitnessCodes{{Created: time.Now()}},
			},
			send: "SendToIndependentWitness",
		},
		actor.TypeCertificateProvider: {
			donor: &donordata.Provided{
				Donor:                    donordata.Donor{FirstNames: "john"},
				CertificateProviderCodes: donordata.WitnessCodes{{Created: time.Now()}},
			},
			send: "SendToCertificateProvider",
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			witnessCodeSender := newMockWitnessCodeSender(t)
			witnessCodeSender.
				On(tc.send, r.Context(), tc.donor, testAppData.Localizer).
				Return(page.ErrTooManyWitnessCodeRequests)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &resendWitnessCodeData{
					App:    testAppData,
					Errors: validation.With("request", validation.CustomError{Label: "pleaseWaitOneMinute"}),
				}).
				Return(nil)

			err := ResendWitnessCode(template.Execute, witnessCodeSender, actorType)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
