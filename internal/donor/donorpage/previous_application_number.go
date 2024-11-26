package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousApplicationNumberData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *previousApplicationNumberForm
}

func PreviousApplicationNumber(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &previousApplicationNumberData{
			App: appData,
			Form: &previousApplicationNumberForm{
				PreviousApplicationNumber: provided.PreviousApplicationNumber,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousApplicationNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.PreviousApplicationNumber = data.Form.PreviousApplicationNumber

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.PreviousApplicationNumber[0] == '7' {
					return donor.PathPreviousFee.Redirect(w, r, appData, provided)
				} else {
					return donor.PathCostOfRepeatApplication.Redirect(w, r, appData, provided)
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
