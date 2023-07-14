package donor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

const (
	peekSize      = 512
	maxUploadSize = 32 << 20 // 32Mb
)

type uploadEvidenceData struct {
	App    page.AppData
	Errors validation.List
}

//go:generate mockery --testonly --inpackage --name S3Client --structname mockS3Client
type S3Client interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func UploadEvidence(tmpl template.Template, donorStore DonorStore, sessionStore SessionStore, s3Client S3Client, bucketName string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &uploadEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+512)
			form := readUploadEvidenceForm(r, sessionStore)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.EvidenceKey = lpa.ID + "-evidence"

				_, err := s3Client.PutObject(r.Context(), &s3.PutObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(lpa.EvidenceKey),
					Body:   bytes.NewReader(form.File),
				})
				if err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.ApplicationReason.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type uploadEvidenceForm struct {
	File  []byte
	Error error
}

func readUploadEvidenceForm(r *http.Request, sessionStore SessionStore) *uploadEvidenceForm {
	reader, err := r.MultipartReader()
	if err != nil {
		return &uploadEvidenceForm{Error: err}
	}

	// csrf part
	{
		csrfSession, err := sessionStore.Get(r, "csrf")
		if err != nil {
			return &uploadEvidenceForm{Error: page.ErrCsrfInvalid}
		}

		cookieValue, ok := csrfSession.Values["token"].(string)
		if !ok {
			return &uploadEvidenceForm{Error: page.ErrCsrfInvalid}
		}

		part, err := reader.NextPart()
		if err != nil {
			return &uploadEvidenceForm{Error: page.ErrCsrfInvalid}
		}

		if part.FormName() != "csrf" {
			return &uploadEvidenceForm{Error: page.ErrCsrfInvalid}
		}

		lmt := io.LimitReader(part, int64(len(cookieValue)+1))
		value, _ := io.ReadAll(lmt)

		if string(value) != cookieValue {
			return &uploadEvidenceForm{Error: page.ErrCsrfInvalid}
		}
	}

	// upload part
	part, err := reader.NextPart()
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}

	if part.FormName() != "upload" {
		return &uploadEvidenceForm{Error: errors.New("unexpected field name")}
	}

	buf := bufio.NewReader(part)

	sniff, _ := buf.Peek(peekSize)
	contentType := http.DetectContentType(sniff)
	if contentType != "application/pdf" {
		return &uploadEvidenceForm{Error: errors.New("unexpected content type")}
	}

	var file bytes.Buffer
	lmt := io.MultiReader(buf, io.LimitReader(part, maxUploadSize-peekSize+1))

	copied, err := io.Copy(&file, lmt)
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}
	if copied > maxUploadSize {
		return &uploadEvidenceForm{Error: errors.New("over size limit")}
	}

	return &uploadEvidenceForm{File: file.Bytes()}
}

func (f *uploadEvidenceForm) Validate() validation.List {
	var errors validation.List

	if f.Error != nil {
		if f.Error == page.ErrCsrfInvalid {
			// this should do
			//   errorHandler(w, r, ErrCsrfInvalid)
			// to match the normal behaviour
			errors.Add("upload", validation.CustomError{Label: "Y"})
		} else {
			errors.Add("upload", validation.CustomError{Label: "X"})
		}
	}

	return errors
}
