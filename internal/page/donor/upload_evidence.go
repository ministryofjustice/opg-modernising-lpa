package donor

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type uploadError int

func (uploadError) Error() string { return "err" }

const (
	peekSize             = 2000     // to account for detecting MS Office files
	maxFileSize          = 32 << 20 // 32Mb
	numberOfAllowedFiles = 5

	errEmptyFile             = uploadError(1)
	errUnexpectedContentType = uploadError(2)
	errFileTooBig            = uploadError(3)
	errTooManyFiles          = uploadError(4)
)

func acceptedMimeTypes() []string {
	return []string{
		"application/pdf",
		"image/png",
		"image/jpeg",
		"image/tiff",
		"image/heic",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.oasis.opendocument.text",
		"application/vnd.oasis.opendocument.spreadsheet",
		"application/vnd.oasis.opendocument.spreadsheet",
	}
}

type uploadEvidenceData struct {
	App                  page.AppData
	Errors               validation.List
	NumberOfAllowedFiles int
	FeeType              page.FeeType
	Evidence             []page.Evidence
	MimeTypes            []string
}

func UploadEvidence(tmpl template.Template, payer Payer, donorStore DonorStore, randomUUID func() string, evidenceBucketName string, s3Client S3Client) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &uploadEvidenceData{
			App:                  appData,
			NumberOfAllowedFiles: numberOfAllowedFiles,
			FeeType:              lpa.FeeType,
			Evidence:             lpa.EvidenceKeys,
			MimeTypes:            acceptedMimeTypes(),
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, numberOfAllowedFiles*maxFileSize+512)
			form := readUploadEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				if form.Action == "upload" {
					for _, file := range form.Files {
						uuid := randomUUID()
						key := lpa.UID + "-evidence-" + uuid

						_, err := s3Client.PutObject(r.Context(), &s3.PutObjectInput{
							Bucket:               aws.String(evidenceBucketName),
							Key:                  aws.String(key),
							Body:                 bytes.NewReader(file.Data),
							ServerSideEncryption: types.ServerSideEncryptionAwsKms,
							Tagging:              aws.String("replicate=true"),
						})
						if err != nil {
							return err
						}

						lpa.EvidenceKeys = append(lpa.EvidenceKeys, page.Evidence{Key: key, Filename: file.Filename})
					}

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}

					data.Evidence = lpa.EvidenceKeys
				} else {
					return payer.Pay(appData, w, r, lpa)
				}
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
	Files  []File
	Action string
	Error  error
}

func readUploadEvidenceForm(r *http.Request) *uploadEvidenceForm {
	form := &uploadEvidenceForm{}

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

	part, err = reader.NextPart()
	if err != nil && err != io.EOF {
		return &uploadEvidenceForm{Error: err}
	}

	if part.FormName() != "action" {
		return &uploadEvidenceForm{Error: errors.New("unexpected field name")}
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(part)
	form.Action = buf.String()

	if form.Action == "upload" {
		// upload part
		files := make([]File, 0, 5)

		for {
			part, err = reader.NextPart()
			if err == io.EOF {
				break
			}

			if err != nil {
				return &uploadEvidenceForm{Error: err}
			}

			if part.FormName() != "upload" {
				return &uploadEvidenceForm{Error: errors.New("unexpected field name")}
			}

			if len(files) >= numberOfAllowedFiles {
				return &uploadEvidenceForm{Error: errTooManyFiles}
			}

			buf := bufio.NewReader(part)
			sniff, _ := buf.Peek(peekSize)

			if len(sniff) == 0 {
				files = append(files, File{Error: errEmptyFile, Filename: part.FileName()})
				continue
			}

			// to account for various docs appearing as zips
			mimetype.SetLimit(0)
			contentType := mimetype.Detect(sniff)

			if !slices.Contains(acceptedMimeTypes(), contentType.String()) {
				files = append(files, File{Error: errUnexpectedContentType, Filename: part.FileName()})
				continue
			}

			var f bytes.Buffer
			lmt := io.MultiReader(buf, io.LimitReader(part, maxFileSize-peekSize+1))

			copied, err := io.Copy(&f, lmt)
			if err != nil && err != io.EOF {
				return &uploadEvidenceForm{Error: err}
			}
			if copied > maxFileSize {
				files = append(files, File{Error: errFileTooBig, Filename: part.FileName()})
				continue
			}

			files = append(files, File{
				Data:     f.Bytes(),
				Filename: part.FileName(),
			})
		}

		form.Files = files
	}

	return form
}

func (f *uploadEvidenceForm) Validate() validation.List {
	var errors validation.List

	if f.Error != nil {
		switch f.Error {
		case errTooManyFiles:
			errors.Add("upload", validation.CustomError{Label: "errorTooManyFiles"})
		default:
			errors.Add("upload", validation.CustomError{Label: "errorGenericUploadProblem"})
		}
	}

	for _, file := range f.Files {
		if file.Error != nil {
			switch file.Error {
			case errUnexpectedContentType:
				errors.Add("upload", validation.FileError{Label: "errorFileIncorrectType", Filename: file.Filename})
			case errFileTooBig:
				errors.Add("upload", validation.FileError{Label: "errorFileTooBig", Filename: file.Filename})
			case errEmptyFile:
				errors.Add("upload", validation.FileError{Label: "errorFileEmpty", Filename: file.Filename})
			default:
				errors.Add("upload", validation.FileError{Label: "errorGenericUploadProblemFile", Filename: file.Filename})
			}
		}
	}

	return errors
}
