package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOneLoginIdentityDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	certificateProvider := &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b"},
		LpaID:            "lpa-id",
	}

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			LpaUID:              "lpa-uid",
			CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"},
			Donor:               lpadata.Donor{FirstNames: "c", LastName: "d"},
		}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &oneLoginIdentityDetailsData{
			App:           testAppData,
			Provided:      certificateProvider,
			DonorFullName: "c d",
		}).
		Return(nil)

	err := OneLoginIdentityDetails(template.Execute, nil, lpaResolvingService)(testAppData, w, r, certificateProvider)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOneLoginIdentityDetailsWhenTemplateErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := OneLoginIdentityDetails(template.Execute, nil, lpaResolvingService)(testAppData, w, r, &certificateproviderdata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostOneLoginIdentityDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	updatedCertificateProvider := &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		LpaID:            "lpa-id",
		DateOfBirth:      date.New("2000", "1", "1"),
		Tasks:            certificateproviderdata.Tasks{ConfirmYourIdentity: task.StateCompleted},
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), updatedCertificateProvider).
		Return(nil)

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"}}, nil)

	err := OneLoginIdentityDetails(nil, certificateProviderStore, lpaResolvingService)(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		DateOfBirth:      date.New("2000", "1", "1"),
		LpaID:            "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenDetailsDoNotMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpadata.CertificateProvider{FirstNames: "x", LastName: "y"}}, nil)

	err := OneLoginIdentityDetails(nil, nil, lpaResolvingService)(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		DateOfBirth:      date.New("2000", "1", "1"),
		LpaID:            "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathProveYourIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	ignoreLpaStoreClient := func() *mockLpaStoreClient { return newMockLpaStoreClient(t) }

	testcases := map[string]struct {
		lpaResolvingService      func() *mockLpaStoreResolvingService
		certificateProviderStore func() *mockCertificateProviderStore
		lpaStoreClient           func() *mockLpaStoreClient
	}{
		"when lpaStoreResolvingService error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(mock.Anything).
					Return(&lpadata.Lpa{}, expectedError)
				return service
			},
			certificateProviderStore: func() *mockCertificateProviderStore {
				return nil
			},
			lpaStoreClient: ignoreLpaStoreClient,
		},
		"when certificateProviderStore.Put() error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(mock.Anything).
					Return(&lpadata.Lpa{CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"}}, nil)
				return service
			},
			certificateProviderStore: func() *mockCertificateProviderStore {
				store := newMockCertificateProviderStore(t)
				store.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(expectedError)
				return store
			},
			lpaStoreClient: ignoreLpaStoreClient,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := OneLoginIdentityDetails(nil, tc.certificateProviderStore(), tc.lpaResolvingService())(testAppData, w, r, &certificateproviderdata.Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1"), Status: identity.StatusConfirmed},
				DateOfBirth:      date.New("2000", "1", "1"),
			})
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
