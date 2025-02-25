package donorpage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadEvidenceSSE(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(document.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).Once()
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(document.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).Once()
	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(document.Documents{
			{Scanned: true},
			{Scanned: true},
		}, nil).Once()

	now := time.Now()

	err := UploadEvidenceSSE(documentStore, nil, 4*time.Millisecond, 2*time.Millisecond, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{})
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
		Return(document.Documents{
			{Scanned: false},
			{Scanned: true},
		}, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(mock.Anything, mock.Anything, mock.Anything)

	err := UploadEvidenceSSE(documentStore, logger, 4*time.Millisecond, 2*time.Millisecond, nil)(testAppData, w, r, &donordata.Provided{})
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
		Return(document.Documents{
			{Scanned: false},
			{Scanned: true},
		}, nil).
		Once()

	documentStore.EXPECT().
		GetAll(r.Context()).
		Return(document.Documents{
			{Scanned: false},
			{Scanned: true},
		}, expectedError).
		Once()

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(mock.Anything, mock.Anything, mock.Anything)

	now := time.Now()

	err := UploadEvidenceSSE(documentStore, logger, 4*time.Millisecond, 2*time.Millisecond, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	bodyBytes, _ := io.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "event: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
}
