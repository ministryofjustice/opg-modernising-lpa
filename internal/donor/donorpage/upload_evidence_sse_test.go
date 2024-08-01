package donorpage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestUploadEvidenceSSE(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).Once()
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).Once()
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: true},
			{Scanned: true},
		}, nil).Once()

	now := time.Now()

	err := UploadEvidenceSSE(documentStore, 4*time.Millisecond, 2*time.Millisecond, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	bodyBytes, _ := io.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "event: message\ndata: {\"finishedScanning\": false, \"scannedCount\": 0}\n\nevent: message\ndata: {\"finishedScanning\": true, \"scannedCount\": 1}\n\nevent: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
}

func TestUploadEvidenceSSEOnDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: false},
			{Scanned: true},
		}, expectedError)

	err := UploadEvidenceSSE(documentStore, 4*time.Millisecond, 2*time.Millisecond, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	bodyBytes, _ := io.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "event: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
}

func TestUploadEvidenceSSEOnDonorStoreErrorWhenRefreshingDocuments(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).
		Once()

	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(page.Documents{
			{Scanned: false},
			{Scanned: true},
		}, expectedError).
		Once()

	now := time.Now()

	err := UploadEvidenceSSE(documentStore, 4*time.Millisecond, 2*time.Millisecond, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	bodyBytes, _ := io.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "event: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
}
