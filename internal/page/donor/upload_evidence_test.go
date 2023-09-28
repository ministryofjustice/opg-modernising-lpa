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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, "", nil)(testAppData, w, r, &page.Lpa{FeeType: page.FullFee})
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

	err := UploadEvidence(template.Execute, nil, nil, nil, "", nil)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidence(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file := addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "who-cares.pdf")

	writer.Close()
	file.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uid"), input.Key) &&
				assert.Equal(t, aws.String("replicate=true"), input.Tagging) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uid"), input.Key) &&
				assert.Equal(t, aws.String("replicate=true"), input.Tagging) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)

	updatedLpa := &page.Lpa{UID: "lpa-uid", EvidenceKeys: []page.Evidence{
		{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf"},
		{Key: "lpa-uid-evidence-a-uid", Filename: "who-cares.pdf"},
	}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, updatedLpa).
		Return(nil)

	err := UploadEvidence(nil, payer, donorStore, func() string { return "a-uid" }, "bucket-name", s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
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
			Errors:               validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"}),
			FeeType:              page.FullFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, "bucket-name", nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: page.FullFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceNumberOfFilesLimitPassed(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file := addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "whatever.pdf")
	file = addFileToUploadField(writer, "whatever.pdf")

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
			Errors:               validation.With("upload", validation.CustomError{Label: "errorTooManyFiles"}),
			FeeType:              page.FullFee,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, "bucket-name", nil)(testAppData, w, r, &page.Lpa{UID: "lpa-uid", FeeType: page.FullFee})
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
			expectedError: validation.FileError{Label: "errorFileEmpty", Filename: "whatever.pdf"},
		},
		"not pdf": {
			fieldName:     "upload",
			fieldContent:  strings.NewReader("I am just text"),
			expectedError: validation.FileError{Label: "errorFileIncorrectType", Filename: "whatever.pdf"},
		},
		"wrong field": {
			fieldName:     "file",
			fieldContent:  bytes.NewReader(dummyData),
			expectedError: validation.CustomError{Label: "errorGenericUploadProblem"},
		},
		"over size pdf": {
			fieldName:     "upload",
			fieldContent:  io.MultiReader(bytes.NewReader(dummyData), randomReader),
			expectedError: validation.FileError{Label: "errorFileTooBig", Filename: "whatever.pdf"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, _ := writer.CreateFormField("csrf")
			io.WriteString(part, "123")

			part, _ = writer.CreateFormFile(tc.fieldName, "whatever.pdf")
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
					Errors:               validation.With("upload", tc.expectedError),
					FeeType:              page.FullFee,
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, nil, nil, "bucket-name", nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", FeeType: page.FullFee})
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

	file := addFileToUploadField(writer, "whatever.pdf")

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, expectedError)

	err := UploadEvidence(nil, nil, nil, func() string { return "a-uid" }, "bucket-name", s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenDonorStoreError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file := addFileToUploadField(writer, "whatever.pdf")

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)

	updatedLpa := &page.Lpa{UID: "lpa-uid", EvidenceKeys: []page.Evidence{
		{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf"},
	}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(expectedError)

	err := UploadEvidence(nil, nil, donorStore, func() string { return "a-uid" }, "bucket-name", s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenPayerError(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file := addFileToUploadField(writer, "whatever.pdf")

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)

	updatedLpa := &page.Lpa{UID: "lpa-uid", EvidenceKeys: []page.Evidence{
		{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf"},
	}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, updatedLpa).
		Return(expectedError)

	err := UploadEvidence(nil, payer, donorStore, func() string { return "a-uid" }, "bucket-name", s3Client)(testAppData, w, r, &page.Lpa{UID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func addFileToUploadField(writer *multipart.Writer, filename string) *os.File {
	file, _ := os.Open("testdata/dummy.pdf")
	part, _ := writer.CreateFormFile("upload", filename)
	io.Copy(part, file)
	return file
}
