package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var now = time.Now()

func TestGetWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Lpa: lpa}).
		Return(nil)

	err := WitnessingYourSignature(template.Execute, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Lpa: lpa}).
		Return(expectedError)

	err := WitnessingYourSignature(template.Execute, nil)(testAppData, w, r, lpa)

	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	testcases := map[string]struct {
		donor    actor.Donor
		methods  []string
		redirect page.LpaPath
	}{
		"can sign": {
			donor:    actor.Donor{CanSign: form.Yes},
			methods:  []string{"SendToCertificateProvider"},
			redirect: page.Paths.WitnessingAsCertificateProvider,
		},
		"cannot sign": {
			donor:    actor.Donor{CanSign: form.No},
			methods:  []string{"SendToCertificateProvider", "SendToIndependentWitness"},
			redirect: page.Paths.WitnessingAsIndependentWitness,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpa := &page.Lpa{
				ID:                    "lpa-id",
				Donor:                 tc.donor,
				DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin},
				CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
			}

			witnessCodeSender := newMockWitnessCodeSender(t)
			for _, method := range tc.methods {
				witnessCodeSender.
					On(method, r.Context(), lpa, mock.Anything).
					Return(nil)
			}

			err := WitnessingYourSignature(nil, witnessCodeSender)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostWitnessingYourSignatureWhenWitnessCodeSenderErrors(t *testing.T) {
	testcases := map[string]func(*mockWitnessCodeSender){
		"SendToCertificateProvider": func(witnessCodeSender *mockWitnessCodeSender) {
			witnessCodeSender.
				On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)
		},
		"SendToIndependentWitness": func(witnessCodeSender *mockWitnessCodeSender) {
			witnessCodeSender.
				On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			witnessCodeSender.
				On("SendToIndependentWitness", mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for name, setupWitnessCodeSender := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpa := &page.Lpa{Donor: actor.Donor{CanSign: form.No}, CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

			witnessCodeSender := newMockWitnessCodeSender(t)
			setupWitnessCodeSender(witnessCodeSender)

			err := WitnessingYourSignature(nil, witnessCodeSender)(testAppData, w, r, lpa)
			assert.Equal(t, expectedError, err)
		})
	}
}
