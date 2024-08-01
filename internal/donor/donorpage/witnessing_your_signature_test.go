package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

	donor := &donordata.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingYourSignatureData{App: testAppData, Donor: donor}).
		Return(nil)

	err := WitnessingYourSignature(template.Execute, nil, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingYourSignatureData{App: testAppData, Donor: donor}).
		Return(expectedError)

	err := WitnessingYourSignature(template.Execute, nil, nil)(testAppData, w, r, donor)

	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donor := &donordata.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 donordata.Donor{CanSign: form.Yes},
		DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:   donordata.CertificateProvider{Mobile: "07535111111"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.EXPECT().
		SendToCertificateProvider(r.Context(), donor, mock.Anything).
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

	donor := &donordata.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 donordata.Donor{CanSign: form.No},
		DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		CertificateProvider:   donordata.CertificateProvider{Mobile: "07535111111"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.EXPECT().
		SendToCertificateProvider(r.Context(), donor, mock.Anything).
		Return(nil)
	witnessCodeSender.EXPECT().
		SendToIndependentWitness(r.Context(), &donordata.DonorProvidedDetails{
			LpaID:                 "lpa-id",
			Donor:                 donordata.Donor{CanSign: form.No},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			CertificateProvider:   donordata.CertificateProvider{Mobile: "07535111111"},
		}, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&donordata.DonorProvidedDetails{
			LpaID:                 "lpa-id",
			Donor:                 donordata.Donor{CanSign: form.No},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			CertificateProvider:   donordata.CertificateProvider{Mobile: "07535111111"},
		}, nil)

	err := WitnessingYourSignature(nil, witnessCodeSender, donorStore)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WitnessingAsIndependentWitness.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingYourSignatureWhenWitnessCodeSenderErrors(t *testing.T) {
	donor := &donordata.DonorProvidedDetails{Donor: donordata.Donor{CanSign: form.No}, CertificateProvider: donordata.CertificateProvider{Mobile: "07535111111"}}

	testcases := map[string]struct {
		setupWitnessCodeSender func(witnessCodeSender *mockWitnessCodeSender)
		setupDonorStore        func(donorStore *mockDonorStore)
	}{
		"SendToCertificateProvider": {
			setupWitnessCodeSender: func(witnessCodeSender *mockWitnessCodeSender) {
				witnessCodeSender.EXPECT().
					SendToCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
			setupDonorStore: func(donorStore *mockDonorStore) {},
		},
		"SendToIndependentWitness": {
			setupWitnessCodeSender: func(witnessCodeSender *mockWitnessCodeSender) {
				witnessCodeSender.EXPECT().
					SendToCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				witnessCodeSender.EXPECT().
					SendToIndependentWitness(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
			setupDonorStore: func(donorStore *mockDonorStore) {
				donorStore.EXPECT().
					Get(mock.Anything).
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
