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

func TestGetConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmYourIdentityData{
			App: testAppData,
			Lpa: &lpadata.Lpa{LpaID: "lpa-id"},
		}).
		Return(nil)

	err := ConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourIdentityWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{LpaID: "lpa-id"})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := ConfirmYourIdentity(nil, nil)(testAppData, w, r, &certificateproviderdata.Provided{
		LpaID: "lpa-id",
		Tasks: certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateInProgress},
	}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathIdentityWithOneLogin.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourIdentityWhenNotStarted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &certificateproviderdata.Provided{
			LpaID: "lpa-id",
			Tasks: certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateInProgress},
		}).
		Return(nil)

	err := ConfirmYourIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathIdentityWithOneLogin.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourIdentityWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	assert.ErrorIs(t, err, expectedError)
}
