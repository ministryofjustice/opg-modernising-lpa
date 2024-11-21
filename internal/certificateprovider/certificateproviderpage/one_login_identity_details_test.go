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

	lpa := &lpadata.Lpa{
		LpaUID:              "lpa-uid",
		CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"},
		Donor:               lpadata.Donor{FirstNames: "c", LastName: "d"},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &oneLoginIdentityDetailsData{
			App:           testAppData,
			Provided:      certificateProvider,
			DonorFullName: "c d",
		}).
		Return(nil)

	err := OneLoginIdentityDetails(template.Execute, nil)(testAppData, w, r, certificateProvider, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOneLoginIdentityDetailsWhenTemplateErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := OneLoginIdentityDetails(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{})
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
		Tasks:            certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
	}

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"}}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), updatedCertificateProvider).
		Return(nil)

	err := OneLoginIdentityDetails(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		DateOfBirth:      date.New("2000", "1", "1"),
		LpaID:            "lpa-id",
	}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenDetailsDoNotMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpadata.CertificateProvider{FirstNames: "x", LastName: "y"}}

	err := OneLoginIdentityDetails(nil, nil)(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		DateOfBirth:      date.New("2000", "1", "1"),
		LpaID:            "lpa-id",
	}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathConfirmYourIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenCertificateProviderStoreErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := OneLoginIdentityDetails(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{
		IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1"), Status: identity.StatusConfirmed},
		DateOfBirth:      date.New("2000", "1", "1"),
	}, &lpadata.Lpa{CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
