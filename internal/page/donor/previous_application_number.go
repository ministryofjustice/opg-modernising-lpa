package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousApplicationNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *previousApplicationNumberForm
}

func PreviousApplicationNumber(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		data := &previousApplicationNumberData{
			App: appData,
			Form: &previousApplicationNumberForm{
				PreviousApplicationNumber: lpa.PreviousApplicationNumber,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousApplicationNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if lpa.PreviousApplicationNumber != data.Form.PreviousApplicationNumber {
					lpa.PreviousApplicationNumber = data.Form.PreviousApplicationNumber

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}
				}

				if lpa.PreviousApplicationNumber[0] == '7' {
					return page.Paths.PreviousFee.Redirect(w, r, appData, lpa)
				} else {
					return page.Paths.EvidenceSuccessfullyUploaded.Redirect(w, r, appData, lpa)
				}
			}
		}

		return tmpl(w, data)
	}
}

type previousApplicationNumberForm struct {
	PreviousApplicationNumber string
}

func readPreviousApplicationNumberForm(r *http.Request) *previousApplicationNumberForm {
	return &previousApplicationNumberForm{
		PreviousApplicationNumber: page.PostFormString(r, "previous-application-number"),
	}
}

func (f *previousApplicationNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("previous-application-number", "previousApplicationNumber", f.PreviousApplicationNumber,
		validation.Empty(),
		validation.ReferenceNumber())

	return errors
}
