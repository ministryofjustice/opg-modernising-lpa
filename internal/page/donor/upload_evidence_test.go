package donor

import (
	"bytes"
	"crypto/rand"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testNow   = time.Now()
	testNowFn = func() time.Time { return testNow }
)

func TestGetUploadEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{{Scanned: false}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			FeeType:              pay.FullFee,
			MimeTypes:            acceptedMimeTypes(),
			Documents:            page.Documents{{Scanned: false}},
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{FeeType: pay.FullFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUploadEvidenceWhenTaskPending(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := UploadEvidence(nil, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: pay.FullFee, Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetUploadEvidenceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{{Scanned: false}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithUploadActionAcceptedFileTypes(t *testing.T) {
	testCases := []string{
		"dummy.docx",
		"dummy.heic",
		"dummy.jpeg",
		"dummy.jpg",
		"dummy.ods",
		"dummy.odt",
		"dummy.pdf",
		"dummy.png",
		"dummy.tif",
		"dummy.xlsx",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			buf, contentType := multipartAction("upload", filename)

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Type", contentType)

			documentStore := newMockDocumentStore(t)
			documentStore.
				On("GetAll", r.Context()).
				Return(page.Documents{}, nil)
			documentStore.
				On("Create", r.Context(), &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee}, filename, mock.Anything).
				Return(page.Document{
					PK:       "LPA#lpa-id",
					SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
					Filename: filename,
					Key:      "lpa-uid/evidence/a-uid",
				}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App: testAppData,
					Documents: page.Documents{{
						PK:       "LPA#lpa-id",
						SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
						Filename: filename,
						Key:      "lpa-uid/evidence/a-uid"},
					},
					NumberOfAllowedFiles: 5,
					MimeTypes:            acceptedMimeTypes(),
					FeeType:              pay.HalfFee,
					StartScan:            "1",
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
			assert.Nil(t, err)
		})
	}
}

func TestPostUploadEvidenceWhenTaskPending(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := UploadEvidence(nil, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: pay.FullFee, Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostUploadEvidenceWithUploadActionMultipleFiles(t *testing.T) {
	buf, contentType := multipartAction("upload", "dummy.pdf", "dummy.png")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)
	documentStore.
		On("Create", r.Context(), &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee}, "dummy.pdf", mock.Anything).
		Return(page.Document{
			PK:       "LPA#lpa-id",
			SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
			Filename: "dummy.pdf",
			Key:      "lpa-uid/evidence/a-uid",
		}, nil).
		Once()
	documentStore.
		On("Create", r.Context(), &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee}, "dummy.png", mock.Anything).
		Return(page.Document{
			PK:       "LPA#lpa-id",
			SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
			Filename: "dummy.png",
			Key:      "lpa-uid/evidence/a-uid",
		}, nil).
		Once()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App: testAppData,
			Documents: page.Documents{
				{
					PK:       "LPA#lpa-id",
					SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
					Filename: "dummy.pdf",
					Key:      "lpa-uid/evidence/a-uid",
				},
				{
					PK:       "LPA#lpa-id",
					SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
					Filename: "dummy.png",
					Key:      "lpa-uid/evidence/a-uid",
				},
			},
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			StartScan:            "1",
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithUploadActionFilenameSpecialCharactersAreEscaped(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "upload")

	file, _ := os.Open("testdata/" + "dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "<img src=1 onerror=alert(document.domain)>’ brute.heic")
	io.Copy(part, file)

	writer.Close()
	file.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)
	documentStore.
		On("Create", r.Context(), &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee}, "&lt;img src=1 onerror=alert(document.domain)&gt;’ brute.heic", mock.Anything).
		Return(page.Document{
			PK:       "LPA#lpa-id",
			SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
			Filename: "&lt;img src=1 onerror=alert(document.domain)&gt;’ brute.heic",
			Key:      "lpa-uid/evidence/a-uid",
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App: testAppData,
			Documents: page.Documents{
				{
					PK:       "LPA#lpa-id",
					SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
					Filename: "&lt;img src=1 onerror=alert(document.domain)&gt;’ brute.heic",
					Key:      "lpa-uid/evidence/a-uid",
				},
			},
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			StartScan:            "1",
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithPayAction(t *testing.T) {
	buf, contentType := multipartAction("pay")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	lpa := &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}

	documents := page.Documents{{
		PK:       "LPA#lpa-id",
		SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
		Filename: "safe.file",
		Key:      "lpa-uid/evidence/a-uid",
		Scanned:  true,
	}, {
		PK:            "LPA#lpa-id",
		SK:            "#DOCUMENT#lpa-uid/evidence/with-virus",
		Filename:      "virus.file",
		Key:           "lpa-uid/evidence/with-virus",
		Scanned:       true,
		VirusDetected: true,
	}}

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(documents, nil)
	documentStore.
		On("Submit", r.Context(), lpa, documents).
		Return(nil)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, lpa).
		Return(nil)

	err := UploadEvidence(nil, payer, documentStore)(testAppData, w, r, lpa)

	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithPayActionWhenPayerError(t *testing.T) {
	buf, contentType := multipartAction("pay")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)
	documentStore.
		On("Submit", r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee}).
		Return(expectedError)

	err := UploadEvidence(nil, payer, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})

	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithPayActionWhenDocumentStoreSubmitErrors(t *testing.T) {
	buf, contentType := multipartAction("pay")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	lpa := &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{{
			PK:       "LPA#lpa-id",
			SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
			Filename: "safe.file",
			Key:      "lpa-uid/evidence/a-uid",
			Scanned:  true,
		}}, nil)
	documentStore.
		On("Submit", r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, lpa)
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithPayActionWhenUnscannedDocument(t *testing.T) {
	buf, contentType := multipartAction("pay")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	lpa := &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Upload}

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{{
			PK:       "LPA#lpa-id",
			SK:       "#DOCUMENT#lpa-uid/evidence/a-uid",
			Filename: "safe.file",
			Key:      "lpa-uid/evidence/a-uid",
		}}, nil)
	documentStore.
		On("Submit", r.Context(), mock.Anything, mock.Anything).
		Return(ErrUnscannedDocumentSubmitted)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *uploadEvidenceData) bool {
			return assert.Equal(t, validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}), data.Errors)
		})).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWithScanResultsActionWithInfectedFiles(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}, nil).
		Once()
	documentStore.
		On("DeleteInfectedDocuments", r.Context(), page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}).
		Return(nil)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "b", VirusDetected: false},
		}, nil).
		Once()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Documents:            page.Documents{{Filename: "b", VirusDetected: false}},
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			Errors:               validation.With("upload", validation.FilesInfectedError{Label: "upload", Filenames: []string{"a", "c", "d"}}),
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithScanResultsActionWithoutInfectedFiles(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: false},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Documents:            page.Documents{{Filename: "a", VirusDetected: false}},
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithPayActionWithInfectedFilesWhenDocumentStoreGetAllErrors(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}, expectedError)

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithScanResultsActionWithInfectedFilesWhenDocumentStoreDeleteInfectedDocumentsError(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}, nil)
	documentStore.
		On("DeleteInfectedDocuments", r.Context(), page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}).
		Return(expectedError)

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithScanResultsActionWithInfectedFilesWhenDocumentStoreGetAllAgainError(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}, nil).
		Once()
	documentStore.
		On("DeleteInfectedDocuments", r.Context(), page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}).
		Return(nil)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "b", VirusDetected: false},
		}, expectedError).
		Once()

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithScanResultsActionWithInfectedFilesWhenTemplateError(t *testing.T) {
	buf, contentType := multipartAction("scanResults")

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", contentType)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}, nil).
		Once()
	documentStore.
		On("DeleteInfectedDocuments", r.Context(), page.Documents{
			{Filename: "a", VirusDetected: true},
			{Filename: "b", VirusDetected: false},
			{Filename: "c", VirusDetected: true},
			{Filename: "d", VirusDetected: true},
		}).
		Return(nil)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Filename: "b", VirusDetected: false},
		}, nil).
		Once()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Documents:            page.Documents{{Filename: "b", VirusDetected: false}},
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			Errors:               validation.With("upload", validation.FilesInfectedError{Label: "upload", Filenames: []string{"a", "c", "d"}}),
		}).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", UID: "lpa-uid", FeeType: pay.HalfFee})

	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenBadCsrfField(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("what")
	io.WriteString(part, "hey")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors:               validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
			FeeType:              pay.FullFee,
			Documents:            page.Documents{},
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: pay.FullFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWhenBadActionField(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("what")
	io.WriteString(part, "hey")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors:               validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
			FeeType:              pay.FullFee,
			Documents:            page.Documents{},
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: pay.FullFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceNumberOfFilesLimitPassed(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "upload")

	file := addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.pdf")

	writer.Close()
	file.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors:               validation.With("upload", validation.CustomError{Label: "errorTooManyFiles"}),
			FeeType:              pay.FullFee,
			Documents:            page.Documents{},
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: pay.FullFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWhenBadUpload(t *testing.T) {
	dummy, _ := os.Open("testdata/dummy.pdf")
	defer dummy.Close()

	dummyData, _ := io.ReadAll(dummy)
	randomReader := io.LimitReader(rand.Reader, int64(maxFileSize))

	testcases := map[string]struct {
		fieldName     string
		fieldContent  io.Reader
		expectedError validation.FormattableError
	}{
		"missing": {
			fieldName:     "upload",
			fieldContent:  strings.NewReader(""),
			expectedError: validation.FileError{Label: "errorFileEmpty", Filename: "dummy.pdf"},
		},
		"not pdf": {
			fieldName:     "upload",
			fieldContent:  strings.NewReader("I am just text"),
			expectedError: validation.FileError{Label: "errorFileIncorrectType", Filename: "dummy.pdf"},
		},
		"wrong field": {
			fieldName:     "file",
			fieldContent:  bytes.NewReader(dummyData),
			expectedError: validation.CustomError{Label: "errorGenericUploadProblem"},
		},
		"over size pdf": {
			fieldName:     "upload",
			fieldContent:  io.MultiReader(bytes.NewReader(dummyData), randomReader),
			expectedError: validation.FileError{Label: "errorFileTooBig", Filename: "dummy.pdf"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, _ := writer.CreateFormField("csrf")
			io.WriteString(part, "123")

			part, _ = writer.CreateFormField("action")
			io.WriteString(part, "upload")

			part, _ = writer.CreateFormFile(tc.fieldName, "dummy.pdf")
			io.Copy(part, tc.fieldContent)

			writer.Close()

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Type", writer.FormDataContentType())

			documentStore := newMockDocumentStore(t)
			documentStore.
				On("GetAll", r.Context()).
				Return(page.Documents{}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App:                  testAppData,
					NumberOfAllowedFiles: 5,
					MimeTypes:            acceptedMimeTypes(),
					Errors:               validation.With("upload", tc.expectedError),
					FeeType:              pay.FullFee,
					Documents:            page.Documents{},
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: pay.FullFee})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetUploadEvidenceDeleteEvidence(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "delete")

	part, _ = writer.CreateFormField("delete")
	io.WriteString(part, "lpa-uid/evidence/a-uid")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.FullFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
			},
			Deleted: "dummy.pdf",
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{})

	assert.Nil(t, err)
}

func TestGetUploadEvidenceDeleteEvidenceWhenUnexpectedFieldName(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "delete")

	part, _ = writer.CreateFormField("not-delete")
	io.WriteString(part, "not-a-key")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.FullFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			},
			Errors: validation.With("delete", validation.CustomError{Label: "errorGenericUploadProblem"}),
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{})

	assert.Nil(t, err)
}

func TestGetUploadEvidenceDeleteEvidenceWhenDocumentStoreDeleteError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "delete")

	part, _ = writer.CreateFormField("delete")
	io.WriteString(part, "lpa-uid/evidence/a-uid")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"}).
		Return(expectedError)

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestGetUploadEvidenceDeleteEvidenceWhenTemplateError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "delete")

	part, _ = writer.CreateFormField("delete")
	io.WriteString(part, "lpa-uid/evidence/a-uid")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.FullFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
			},
			Deleted: "dummy.pdf",
		}).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWithCloseConnectionAction(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "closeConnection")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			},
			Errors: validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: pay.HalfFee})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWithCloseConnectionActionWhenDocumentStoreDeleteError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "closeConnection")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false}).
		Return(expectedError)

	err := UploadEvidence(nil, nil, documentStore)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: pay.HalfFee})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWithCloseConnectionActionWhenTemplateError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "closeConnection")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			},
			Errors: validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
		}).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: pay.HalfFee})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWithCancelUploadAction(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "cancelUpload")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", r.Context()).
		Return(page.Documents{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false},
		}, nil)
	documentStore.
		On("Delete", r.Context(), page.Document{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png", Scanned: false}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              pay.HalfFee,
			Documents: page.Documents{
				{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf", Scanned: true},
			},
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, documentStore)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: pay.HalfFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func multipartAction(action string, files ...string) (bytes.Buffer, string) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, action)

	for _, file := range files {
		addFileToUploadField(writer, file)
	}

	writer.Close()
	return buf, writer.FormDataContentType()
}

func addFileToUploadField(writer *multipart.Writer, filename string) *os.File {
	file, _ := os.Open("testdata/" + filename)
	defer file.Close()

	part, _ := writer.CreateFormFile("upload", filename)
	io.Copy(part, file)

	return file
}
