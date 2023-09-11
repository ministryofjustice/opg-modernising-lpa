package donor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type uploadError int

func (uploadError) Error() string { return "err" }

const (
	peekSize      = 512
	maxUploadSize = 32 << 20 // 32Mb

	errUploadMissing         = uploadError(1)
	errUnexpectedContentType = uploadError(2)
	errUploadTooBig          = uploadError(3)
)

type uploadEvidenceData struct {
	App    page.AppData
	Errors validation.List
}

//go:generate mockery --testonly --inpackage --name S3Client --structname mockS3Client
type S3Client interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func UploadEvidence(tmpl template.Template, payer Payer) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &uploadEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			return payer.Pay(appData, w, r, lpa)
		}

		return tmpl(w, data)
	}
}

func UploadEvidenceAjax(donorStore DonorStore, s3Client S3Client, bucketName string, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &uploadEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+512)
			form := readUploadEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {

				key := lpa.UID + "-evidence-" + now().Format(time.RFC3339Nano)
				log.Print(key)
				lpa.EvidenceKeys.Add(key)

				_, err := s3Client.PutObject(r.Context(), &s3.PutObjectInput{
					Bucket:               aws.String(bucketName),
					Key:                  aws.String(key),
					Body:                 bytes.NewReader(form.Files),
					ServerSideEncryption: types.ServerSideEncryptionAwsKms,
				})
				if err != nil {
					return err
				}

				return donorStore.Put(r.Context(), lpa)
			}

			return nil
		}

		return nil
	}
}

type uploadEvidenceForm struct {
	Files []byte
	Error error
}

func readUploadEvidenceForm(r *http.Request) *uploadEvidenceForm {
	reader, err := r.MultipartReader()
	if err != nil {
		return &uploadEvidenceForm{Error: err}
	}

	// upload part
	part, err := reader.NextPart()
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}

	if part.FormName() != "documents" {
		return &uploadEvidenceForm{Error: errors.New("unexpected field name")}
	}

	buf := bufio.NewReader(part)

	sniff, _ := buf.Peek(peekSize)
	if len(sniff) == 0 {
		return &uploadEvidenceForm{Error: errUploadMissing}
	}

	contentType := http.DetectContentType(sniff)
	if contentType != "application/pdf" {
		return &uploadEvidenceForm{Error: errUnexpectedContentType}
	}

	var file bytes.Buffer
	lmt := io.MultiReader(buf, io.LimitReader(part, maxUploadSize-peekSize+1))

	copied, err := io.Copy(&file, lmt)
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}
	if copied > maxUploadSize {
		return &uploadEvidenceForm{Error: errUploadTooBig}
	}

	return &uploadEvidenceForm{Files: file.Bytes()}
}

func (f *uploadEvidenceForm) Validate() validation.List {
	var errors validation.List

	if f.Error != nil {
		switch f.Error {
		case errUploadMissing:
			errors.Add("upload", validation.CustomError{Label: "errorUploadMissing"})
		case errUnexpectedContentType:
			errors.Add("upload", validation.CustomError{Label: "errorFileIncorrectType"})
		case errUploadTooBig:
			errors.Add("upload", validation.CustomError{Label: "errorFileTooBig"})
		default:
			errors.Add("upload", validation.CustomError{Label: "errorGenericUploadProblem"})
		}
	}

	return errors
}
