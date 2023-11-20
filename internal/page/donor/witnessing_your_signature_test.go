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

	donor := &actor.DonorProvidedDetails{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Donor: donor}).
		Return(nil)

	err := WitnessingYourSignature(template.Execute, nil, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Donor: donor}).
		Return(expectedError)

	err := WitnessingYourSignature(template.Execute, nil, nil)(testAppData, w, r, donor)

	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donor := &actor.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 actor.Donor{CanSign: form.Yes},
		DonorIdentityUserData: identity.UserData{OK: true},
		CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("SendToCertificateProvider", r.Context(), donor, mock.Anything).
		Return(nil)

	err := WitnessingYourSignature(nil, witnessCodeSender, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WitnessingAsCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingYourSignatureCannotSign(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donor := &actor.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 actor.Donor{CanSign: form.No},
		DonorIdentityUserData: identity.UserData{OK: true},
		CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("SendToCertificateProvider", r.Context(), donor, mock.Anything).
		Return(nil)
	witnessCodeSender.
		On("SendToIndependentWitness", r.Context(), &actor.DonorProvidedDetails{
			LpaID:                 "lpa-id",
			Donor:                 actor.Donor{CanSign: form.No},
			DonorIdentityUserData: identity.UserData{OK: true},
			CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
		}, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&actor.DonorProvidedDetails{
			LpaID:                 "lpa-id",
			Donor:                 actor.Donor{CanSign: form.No},
			DonorIdentityUserData: identity.UserData{OK: true},
			CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
		}, nil)

	err := WitnessingYourSignature(nil, witnessCodeSender, donorStore)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WitnessingAsIndependentWitness.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingYourSignatureWhenWitnessCodeSenderErrors(t *testing.T) {
	donor := &actor.DonorProvidedDetails{Donor: actor.Donor{CanSign: form.No}, CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	testcases := map[string]struct {
		setupWitnessCodeSender func(witnessCodeSender *mockWitnessCodeSender)
		setupDonorStore        func(donorStore *mockDonorStore)
	}{
		"SendToCertificateProvider": {
			setupWitnessCodeSender: func(witnessCodeSender *mockWitnessCodeSender) {
				witnessCodeSender.
					On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
			setupDonorStore: func(donorStore *mockDonorStore) {},
		},
		"SendToIndependentWitness": {
			setupWitnessCodeSender: func(witnessCodeSender *mockWitnessCodeSender) {
				witnessCodeSender.
					On("SendToCertificateProvider", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				witnessCodeSender.
					On("SendToIndependentWitness", mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
			setupDonorStore: func(donorStore *mockDonorStore) {
				donorStore.
					On("Get", mock.Anything, mock.Anything, mock.Anything).
					Return(donor, nil)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			witnessCodeSender := newMockWitnessCodeSender(t)
			tc.setupWitnessCodeSender(witnessCodeSender)

			donorStore := newMockDonorStore(t)
			tc.setupDonorStore(donorStore)

			err := WitnessingYourSignature(nil, witnessCodeSender, donorStore)(testAppData, w, r, donor)
			assert.Equal(t, expectedError, err)
		})
	}
}
