package donor

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"

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

func UploadEvidence(tmpl template.Template, payer Payer, donorStore DonorStore, randomUUID func() string, evidenceBucketName string, s3Client S3Client) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &uploadEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+512)
			form := readUploadEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				for _, file := range form.Files {
					uuid := randomUUID()
					key := lpa.UID + "-evidence-" + uuid

					_, err := s3Client.PutObject(r.Context(), &s3.PutObjectInput{
						Bucket:               aws.String(evidenceBucketName),
						Key:                  aws.String(key),
						Body:                 bytes.NewReader(file.Data),
						ServerSideEncryption: types.ServerSideEncryptionAwsKms,
					})
					if err != nil {
						return err
					}

					lpa.EvidenceKeys = append(lpa.EvidenceKeys, page.Evidence{Key: key, Filename: file.Filename})
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return payer.Pay(appData, w, r, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type File struct {
	Data     []byte
	Filename string
	Error    error
}

type uploadEvidenceForm struct {
	Files []File
	Error error
}

func readUploadEvidenceForm(r *http.Request) *uploadEvidenceForm {
	reader, err := r.MultipartReader()
	if err != nil {
		return &uploadEvidenceForm{Error: err}
	}

	// first part will be csrf, so skip
	part, err := reader.NextPart()
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}

	if part.FormName() != "csrf" {
		return &uploadEvidenceForm{Error: errors.New("unexpected field name")}
	}

	// upload part
	var files []File
	var formLevelError error
	for {
		part, err = reader.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			formLevelError = err
			break
		}

		if part.FormName() != "upload" {
			formLevelError = errors.New("unexpected field name")
			break
		}

		buf := bufio.NewReader(part)

		sniff, _ := buf.Peek(peekSize)
		if len(sniff) == 0 {
			formLevelError = errUploadMissing
			break
		}

		contentType := http.DetectContentType(sniff)
		if contentType != "application/pdf" {
			files = append(files, File{Error: errUnexpectedContentType, Filename: part.FileName()})
			continue
		}

		var f bytes.Buffer
		lmt := io.MultiReader(buf, io.LimitReader(part, maxUploadSize-peekSize+1))

		_, err := io.Copy(&f, lmt)

		if err != nil {
			file := File{Error: err, Filename: part.FileName()}

			if errors.As(err, new(*http.MaxBytesError)) {
				file.Error = errUploadTooBig
			}

			files = append(files, file)
			break
		}

		files = append(files, File{
			Data:     f.Bytes(),
			Filename: part.FileName(),
		})
	}

	return &uploadEvidenceForm{Files: files, Error: formLevelError}
}

func (f *uploadEvidenceForm) Validate() validation.List {
	var errors validation.List

	if f.Error != nil {
		switch f.Error {
		case errUploadMissing:
			errors.Add("upload", validation.CustomError{Label: "errorUploadMissing"})
		default:
			errors.Add("upload", validation.CustomError{Label: "errorGenericUploadProblem"})
		}
	}

	for _, file := range f.Files {
		if file.Error != nil {
			switch file.Error {
			case errUnexpectedContentType:
				errors.Add("upload", validation.FileError{Label: "errorFileIncorrectType", Filename: file.Filename})
			case errUploadTooBig:
				errors.Add("upload", validation.FileError{Label: "errorFileTooBig", Filename: file.Filename})
			default:
				errors.Add("upload", validation.FileError{Label: "errorGenericUploadProblemFile", Filename: file.Filename})
			}
		}
	}

	return errors
}
