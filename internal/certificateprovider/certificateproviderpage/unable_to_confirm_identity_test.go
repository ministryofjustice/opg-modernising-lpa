package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUnableToConfirmIdentity(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &unableToConfirmIdentityData{
			App:   testAppData,
			Donor: lpadata.Donor{FirstNames: "a"},
		}).
		Return(nil)

	err := UnableToConfirmIdentity(template.Execute, nil)(testAppData, w, r, nil, &lpadata.Lpa{Donor: lpadata.Donor{FirstNames: "a"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUnableToConfirmIdentityWhenTemplateErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := UnableToConfirmIdentity(template.Execute, nil)(testAppData, w, r, nil, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUnableToConfirmIdentity(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &certificateproviderdata.Provided{
			LpaID: "lpa-id",
			Tasks: certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
		}).
		Return(nil)

	err := UnableToConfirmIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostUnableToConfirmIdentityWhenCertificateProviderStoreErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := UnableToConfirmIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
