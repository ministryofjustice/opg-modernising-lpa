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
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUploadEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App: testAppData,
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, nil, nil, "")(testAppData, w, r, &page.Lpa{})
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

	err := UploadEvidence(template.Execute, nil, nil, nil, "")(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidence(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file, _ := os.Open("testdata/dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "whatever.pdf")
	io.Copy(part, file)

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[any]any{"token": "123"}}, nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("evidence-bucket"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-id-evidence"), input.Key)
		})).
		Return(nil, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{ID: "lpa-id", EvidenceKey: "lpa-id-evidence"}).
		Return(nil)

	err := UploadEvidence(nil, donorStore, sessionStore, s3Client, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ApplicationReason.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostUploadEvidenceWhenBadCsrf(t *testing.T) {
	testcases := map[string]struct {
		fieldName    string
		fieldContent string
	}{
		"bad value": {
			fieldName:    "csrf",
			fieldContent: "456",
		},
		"wrong field": {
			fieldName:    "not-csrf",
			fieldContent: "123",
		},
		"over size value": {
			fieldName:    "csrf",
			fieldContent: "4567",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, _ := writer.CreateFormField(tc.fieldName)
			io.WriteString(part, tc.fieldContent)

			writer.Close()

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", &buf)
			r.Header.Set("Content-Type", writer.FormDataContentType())

			sessionStore := newMockSessionStore(t)
			sessionStore.
				On("Get", r, "csrf").
				Return(&sessions.Session{Values: map[any]any{"token": "123"}}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App:    testAppData,
					Errors: validation.With("upload", validation.CustomError{Label: "Y"}),
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, sessionStore, nil, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostUploadEvidenceWhenBadUpload(t *testing.T) {
	dummy, _ := os.Open("testdata/dummy.pdf")
	defer dummy.Close()

	dummyData, _ := io.ReadAll(dummy)
	randomReader := io.LimitReader(rand.Reader, int64(maxUploadSize-len(dummyData)+1))

	testcases := map[string]struct {
		fieldName    string
		fieldContent io.Reader
	}{
		"not pdf": {
			fieldName:    "upload",
			fieldContent: strings.NewReader("I am just text"),
		},
		"wrong field": {
			fieldName:    "file",
			fieldContent: bytes.NewReader(dummyData),
		},
		"over size pdf": {
			fieldName:    "upload",
			fieldContent: io.MultiReader(bytes.NewReader(dummyData), randomReader),
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

			sessionStore := newMockSessionStore(t)
			sessionStore.
				On("Get", r, "csrf").
				Return(&sessions.Session{Values: map[any]any{"token": "123"}}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &uploadEvidenceData{
					App:    testAppData,
					Errors: validation.With("upload", validation.CustomError{Label: "X"}),
				}).
				Return(nil)

			err := UploadEvidence(template.Execute, nil, sessionStore, nil, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostUploadEvidenceWhenSessionStoreErrors(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file, _ := os.Open("testdata/dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "whatever.pdf")
	io.Copy(part, file)

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(nil, expectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &uploadEvidenceData{
			App:    testAppData,
			Errors: validation.With("upload", validation.CustomError{Label: "Y"}),
		}).
		Return(nil)

	err := UploadEvidence(template.Execute, nil, sessionStore, nil, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUploadEvidenceWhenS3ClientErrors(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file, _ := os.Open("testdata/dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "whatever.pdf")
	io.Copy(part, file)

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[any]any{"token": "123"}}, nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := UploadEvidence(nil, nil, sessionStore, s3Client, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostUploadEvidenceWhenDonorStoreErrors(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormField("csrf")
	io.WriteString(part, "123")

	file, _ := os.Open("testdata/dummy.pdf")
	part, _ = writer.CreateFormFile("upload", "whatever.pdf")
	io.Copy(part, file)

	file.Close()
	writer.Close()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", &buf)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "csrf").
		Return(&sessions.Session{Values: map[any]any{"token": "123"}}, nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", r.Context(), mock.Anything).
		Return(nil, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := UploadEvidence(nil, donorStore, sessionStore, s3Client, "evidence-bucket")(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
