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

func TestGetReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readTheLpaData{App: testAppData, Lpa: donor}).
		Return(nil)

	err := ReadTheLpa(template.Execute, nil)(testAppData, w, r, nil, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, nil)(testAppData, w, r, nil, nil)

	assert.Equal(t, expectedError, err)
}

func TestPostReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &certificateproviderdata.Provided{
			Tasks: certificateproviderdata.Tasks{
				ReadTheLpa: task.StateCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{
		LpaID:                            "lpa-id",
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Paid:                             true,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostReadTheLpaWithAttorneyWhenCertificateStorePutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{
		LpaID:                            "lpa-id",
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Paid:                             true,
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
