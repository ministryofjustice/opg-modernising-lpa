package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourDetails(t *testing.T) {
	testcases := map[string]struct {
		DonorChannel                    actor.Channel
		PhoneNumberLabel                string
		CertificateProviderRelationship actor.CertificateProviderRelationship
		AddressLabel                    string
		DetailsComponentContent         string
	}{
		"online donor": {
			DonorChannel:            actor.ChannelOnline,
			PhoneNumberLabel:        "mobileNumber",
			AddressLabel:            "address",
			DetailsComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
		},
		"paper donor": {
			DonorChannel:            actor.ChannelPaper,
			PhoneNumberLabel:        "contactNumber",
			AddressLabel:            "address",
			DetailsComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
		},
		"lay CP": {
			CertificateProviderRelationship: actor.Personally,
			AddressLabel:                    "address",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
			PhoneNumberLabel:                "mobileNumber",
		},
		"professional CP": {
			CertificateProviderRelationship: actor.Professionally,
			AddressLabel:                    "workAddress",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional",
			PhoneNumberLabel:                "mobileNumber",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpa := &lpastore.Lpa{
				Donor:               lpastore.Donor{Channel: tc.DonorChannel},
				CertificateProvider: lpastore.CertificateProvider{Relationship: tc.CertificateProviderRelationship},
			}
			certificateProvider := &actor.CertificateProviderProvidedDetails{}

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(lpa, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &confirmYourDetailsData{
					App:                    testAppData,
					Lpa:                    lpa,
					CertificateProvider:    certificateProvider,
					PhoneNumberLabel:       tc.PhoneNumberLabel,
					AddressLabel:           tc.AddressLabel,
					DetailComponentContent: tc.DetailsComponentContent,
				}).
				Return(nil)

			err := ConfirmYourDetails(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, certificateProvider)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmYourDetailsWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := ConfirmYourDetails(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	testCases := map[string]struct {
		signedAt time.Time
		redirect page.CertificateProviderPath
	}{
		"signed":     {signedAt: time.Now(), redirect: page.Paths.CertificateProvider.TaskList},
		"not signed": {redirect: page.Paths.CertificateProvider.YourRole},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.Lpa{SignedAt: tc.signedAt}, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Put(r.Context(), &actor.CertificateProviderProvidedDetails{
					LpaID: "lpa-id",
					Tasks: actor.CertificateProviderTasks{
						ConfirmYourDetails: actor.TaskCompleted,
					},
				}).
				Return(nil)

			err := ConfirmYourDetails(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r, &actor.CertificateProviderProvidedDetails{})
	assert.Equal(t, expectedError, err)
}
