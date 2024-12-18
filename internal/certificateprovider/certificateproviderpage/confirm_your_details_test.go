package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourDetails(t *testing.T) {
	testcases := map[string]struct {
		DonorChannel                    lpadata.Channel
		PhoneNumberLabel                string
		CertificateProviderRelationship lpadata.CertificateProviderRelationship
		AddressLabel                    string
		DetailsComponentContent         string
		ShowPhone                       bool
		ShowHomeAddress                 bool
		PhoneNumber                     string
	}{
		"online donor": {
			DonorChannel:            lpadata.ChannelOnline,
			PhoneNumberLabel:        "mobileNumber",
			AddressLabel:            "address",
			DetailsComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
			PhoneNumber:             "123",
			ShowPhone:               true,
		},
		"paper donor": {
			DonorChannel:            lpadata.ChannelPaper,
			PhoneNumberLabel:        "contactNumber",
			AddressLabel:            "address",
			DetailsComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
			PhoneNumber:             "123",
			ShowPhone:               true,
			ShowHomeAddress:         true,
		},
		"lay CP": {
			CertificateProviderRelationship: lpadata.Personally,
			AddressLabel:                    "address",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
			PhoneNumberLabel:                "mobileNumber",
			PhoneNumber:                     "123",
			ShowPhone:                       true,
		},
		"professional CP": {
			CertificateProviderRelationship: lpadata.Professionally,
			AddressLabel:                    "workAddress",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional",
			PhoneNumberLabel:                "mobileNumber",
			PhoneNumber:                     "123",
			ShowPhone:                       true,
			ShowHomeAddress:                 true,
		},
		"missing phone": {
			CertificateProviderRelationship: lpadata.Personally,
			AddressLabel:                    "address",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLayMissingPhone",
			PhoneNumberLabel:                "mobileNumber",
		},
		"professional missing phone": {
			CertificateProviderRelationship: lpadata.Professionally,
			AddressLabel:                    "workAddress",
			DetailsComponentContent:         "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessionalMissingPhone",
			PhoneNumberLabel:                "mobileNumber",
			ShowHomeAddress:                 true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpa := &lpadata.Lpa{
				Donor: lpadata.Donor{Channel: tc.DonorChannel},
				CertificateProvider: lpadata.CertificateProvider{
					Relationship: tc.CertificateProviderRelationship,
					Phone:        tc.PhoneNumber,
				},
			}
			certificateProvider := &certificateproviderdata.Provided{}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &confirmYourDetailsData{
					App:                    testAppData,
					Lpa:                    lpa,
					CertificateProvider:    certificateProvider,
					PhoneNumberLabel:       tc.PhoneNumberLabel,
					AddressLabel:           tc.AddressLabel,
					DetailComponentContent: tc.DetailsComponentContent,
					ShowPhone:              tc.ShowPhone,
					ShowHomeAddress:        tc.ShowHomeAddress,
				}).
				Return(nil)

			err := ConfirmYourDetails(template.Execute, nil)(testAppData, w, r, certificateProvider, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	testCases := map[string]struct {
		signedAt time.Time
		redirect certificateprovider.Path
	}{
		"signed":     {signedAt: time.Now(), redirect: certificateprovider.PathTaskList},
		"not signed": {redirect: certificateprovider.PathYourRole},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Put(r.Context(), &certificateproviderdata.Provided{
					LpaID: "lpa-id",
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourDetails: task.StateCompleted,
					},
				}).
				Return(nil)

			err := ConfirmYourDetails(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{SignedAt: tc.signedAt, WitnessedByCertificateProviderAt: tc.signedAt})
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

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}
