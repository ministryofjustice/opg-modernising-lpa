package donor

import (
	"bufio"
	"bytes"
	"errors"
	"html"
	"io"
	"net/http"
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type uploadError int

func (uploadError) Error() string { return "err" }

const (
	peekSize             = 2000     // to account for detecting MS Office files
	maxFileSize          = 20 << 20 // 20Mb
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
	Documents            page.Documents
	MimeTypes            []string
	Deleted              string
	StartScan            string
}

func UploadEvidence(tmpl template.Template, payer Payer, randomUUID func() string, documentStore DocumentStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if lpa.Tasks.PayForLpa.IsPending() {
			return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
		}

		documents, err := documentStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		data := &uploadEvidenceData{
			App:                  appData,
			NumberOfAllowedFiles: numberOfAllowedFiles,
			FeeType:              lpa.FeeType,
			Documents:            documents,
			MimeTypes:            acceptedMimeTypes(),
		}

		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, numberOfAllowedFiles*maxFileSize+512)
			form := readUploadEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				switch form.Action {
				case "upload":
					var uploadedDocuments []page.Document

					for _, file := range form.Files {
						key := lpa.UID + "/evidence/" + randomUUID()

						document := page.Document{
							PK:       "LPA#" + lpa.ID,
							SK:       "#DOCUMENT#" + key,
							Filename: file.Filename,
							Key:      key,
						}
						if err := documentStore.Put(r.Context(), document, file.Data); err != nil {
							return err
						}

						uploadedDocuments = append(uploadedDocuments, document)
					}

					data.Documents = uploadedDocuments
					data.StartScan = "1"

				case "scanResults":
					infectedFilenames := documents.InfectedFilenames()

					if len(infectedFilenames) > 0 {
						if err := documentStore.DeleteInfectedDocuments(r.Context(), documents); err != nil {
							return err
						}

						refreshedDocuments, err := documentStore.GetAll(r.Context())
						if err != nil {
							return err
						}

						data.Errors = validation.With("upload", validation.FilesInfectedError{Label: "upload", Filenames: infectedFilenames})
						data.Documents = refreshedDocuments

						return tmpl(w, data)
					}

				case "pay":
					return payer.Pay(appData, w, r, lpa)

				case "delete":
					document := documents.Get(form.DeleteKey)
					if document.Key != "" {
						data.Deleted = document.Filename

						if err := documentStore.Delete(r.Context(), document); err != nil {
							return err
						}
						documents.Delete(document.Key)

						data.Documents = documents
					}

				case "closeConnection", "cancelUpload":
					for _, d := range documents {
						if d.Key != "" && !d.Scanned {
							if err := documentStore.Delete(r.Context(), d); err != nil {
								return err
							}

							documents.Delete(d.Key)
						}
					}

					data.Documents = documents

					if form.Action == "closeConnection" {
						data.Errors = validation.With("upload", validation.CustomError{Label: "errorGenericUploadProblem"})
					}

					return tmpl(w, data)

				default:
					return errors.New("unexpected action")
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
	Files     []File
	Action    string
	DeleteKey string
	Error     error
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
				return &uploadEvidenceForm{Error: errors.New("unexpected field name"), Action: "upload"}
			}

			if len(files) >= numberOfAllowedFiles {
				return &uploadEvidenceForm{Error: errTooManyFiles, Action: "upload"}
			}

			buf := bufio.NewReader(part)
			sniff, _ := buf.Peek(peekSize)

			filename := html.EscapeString(part.FileName())

			if len(sniff) == 0 {
				files = append(files, File{Error: errEmptyFile, Filename: filename})
				continue
			}

			mimetype.SetLimit(peekSize)
			contentType := mimetype.Detect(sniff)

			if !slices.Contains(acceptedMimeTypes(), contentType.String()) {
				files = append(files, File{Error: errUnexpectedContentType, Filename: filename})
				continue
			}

			var f bytes.Buffer
			lmt := io.MultiReader(buf, io.LimitReader(part, maxFileSize-peekSize+1))

			copied, err := io.Copy(&f, lmt)
			if err != nil && err != io.EOF {
				return &uploadEvidenceForm{Error: err, Action: "upload"}
			}
			if copied > maxFileSize {
				files = append(files, File{Error: errFileTooBig, Filename: filename})
				continue
			}

			files = append(files, File{
				Data:     f.Bytes(),
				Filename: filename,
			})
		}

		form.Files = files
	}

	if form.Action == "delete" {
		part, err = reader.NextPart()

		if part.FormName() != "delete" {
			return &uploadEvidenceForm{Error: errors.New("unexpected field name"), Action: "delete"}
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(part)
		form.DeleteKey = buf.String()
	}

	return form
}

func (f *uploadEvidenceForm) Validate() validation.List {
	var errors validation.List

	if f.Error != nil {
		field := "upload"

		if f.Action != "" {
			field = f.Action
		}

		switch f.Error {
		case errTooManyFiles:
			errors.Add("upload", validation.CustomError{Label: "errorTooManyFiles"})
		default:
			errors.Add(field, validation.CustomError{Label: "errorGenericUploadProblem"})
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
