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

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUploadEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			FeeType:              page.FullFee,
			MimeTypes:            acceptedMimeTypes(),
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{FeeType: page.FullFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUploadEvidenceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
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
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, _ := writer.CreateFormField("csrf")
			io.WriteString(part, "123")

			part, _ = writer.CreateFormField("action")
			io.WriteString(part, "upload")

			file := addFileToUploadField(writer, filename)

			writer.Close()
			file.Close()

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Type", writer.FormDataContentType())

			s3Client := newMockS3Client(t)
			s3Client.
				On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
				Return(nil)

			evidence := []page.Evidence{
				{Key: "lpa-uid/evidence/a-uid", Filename: filename},
			}
			updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), updatedLpa).
				Return(nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App:                  testAppData,
					Evidence:             evidence,
					NumberOfAllowedFiles: 5,
					MimeTypes:            acceptedMimeTypes(),
					FeeType:              page.HalfFee,
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee})
			assert.Nil(t, err)
		})
	}
}

func TestPostUploadEvidenceWithUploadActionMultipleFiles(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "upload")

	file := addFileToUploadField(writer, "dummy.pdf")
	file = addFileToUploadField(writer, "dummy.png")

	writer.Close()
	file.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
		Return(nil)
	s3Client.
		On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
		Return(nil)

	evidence := []page.Evidence{
		{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
		{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.png"},
	}
	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Evidence:             evidence,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              page.HalfFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee})
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

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
		Return(nil)

	evidence := []page.Evidence{
		{Key: "lpa-uid/evidence/a-uid", Filename: "&lt;img src=1 onerror=alert(document.domain)&gt;’ brute.heic"},
	}
	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Evidence:             evidence,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              page.HalfFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee})
	assert.Nil(t, err)
}

func TestPostUploadEvidenceWithPayAction(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "pay")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee}).
		Return(nil)

	err := UploadEvidence(nil, payer, nil, nil, nil)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee})
	assert.Nil(t, err)
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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors: validation.With(""+
				"upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
			FeeType: page.FullFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: page.FullFee})
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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors:               validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
			FeeType:              page.FullFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: page.FullFee})
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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			Errors:               validation.With("upload", validation.CustomError{Label: "errorTooManyFiles"}),
			FeeType:              page.FullFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.FullFee})
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

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App:                  testAppData,
					NumberOfAllowedFiles: 5,
					MimeTypes:            acceptedMimeTypes(),
					Errors:               validation.With("upload", tc.expectedError),
					FeeType:              page.FullFee,
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: page.FullFee})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostUploadEvidenceWhenS3ClientErrors(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "upload")

	file := addFileToUploadField(writer, "dummy.pdf")

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
		Return(expectedError)

	err := UploadEvidence(nil, nil, nil, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenDonorStoreError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "upload")

	file := addFileToUploadField(writer, "dummy.pdf")

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), "lpa-uid/evidence/a-uid", mock.Anything).
		Return(nil)

	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: []page.Evidence{
		{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
	}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(expectedError)

	err := UploadEvidence(nil, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenPayerError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	part, _ = writer.CreateFormField("action")
	io.WriteString(part, "pay")

	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee}).
		Return(expectedError)

	err := UploadEvidence(nil, payer, nil, nil, nil)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.HalfFee})
	assert.Equal(t, expectedError, err)
}

func TestGetUploadEvidenceDeleteEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?delete=lpa-uid/evidence/a-uid", nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", r.Context(), "lpa-uid/evidence/a-uid").
		Return(nil)

	evidence := []page.Evidence{
		{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
	}
	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Evidence:             evidence,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              page.HalfFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{
		UID:     "lpa-uid",
		FeeType: page.HalfFee,
		Evidence: []page.Evidence{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		},
	})
	assert.Nil(t, err)
}

func TestGetUploadEvidenceDeleteEvidenceWhenS3ClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?delete=lpa-uid/evidence/a-uid", nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", r.Context(), "lpa-uid/evidence/a-uid").
		Return(expectedError)

	err := UploadEvidence(nil, nil, nil, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{
		UID:     "lpa-uid",
		FeeType: page.HalfFee,
		Evidence: []page.Evidence{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestGetUploadEvidenceDeleteEvidenceOnDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?delete=lpa-uid/evidence/a-uid", nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", r.Context(), "lpa-uid/evidence/a-uid").
		Return(nil)

	evidence := []page.Evidence{
		{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
	}
	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(expectedError)

	err := UploadEvidence(nil, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{
		UID:     "lpa-uid",
		FeeType: page.HalfFee,
		Evidence: []page.Evidence{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func TestGetUploadEvidenceDeleteEvidenceOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?delete=lpa-uid/evidence/a-uid", nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("DeleteObject", r.Context(), "lpa-uid/evidence/a-uid").
		Return(nil)

	evidence := []page.Evidence{
		{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
	}
	updatedLpa := &page.Lpa{UID: "lpa-uid", Evidence: evidence, FeeType: page.HalfFee}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:                  testAppData,
			Evidence:             evidence,
			NumberOfAllowedFiles: 5,
			MimeTypes:            acceptedMimeTypes(),
			FeeType:              page.HalfFee,
		}).
		Return(expectedError)

	err := UploadEvidence(template.Execute, nil, donorStore, func() string { return "a-uid" }, s3Client)(testAppData, w, r, &page.Lpa{
		UID:     "lpa-uid",
		FeeType: page.HalfFee,
		Evidence: []page.Evidence{
			{Key: "lpa-uid/evidence/a-uid", Filename: "dummy.pdf"},
			{Key: "lpa-uid/evidence/another-uid", Filename: "dummy.png"},
		},
	})

	assert.Equal(t, expectedError, err)
}

func addFileToUploadField(writer *multipart.Writer, filename string) *os.File {
	file, _ := os.Open("testdata/" + filename)
	part, _ := writer.CreateFormFile("upload", filename)
	io.Copy(part, file)
	return file
}
