package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type previousApplicationNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *previousApplicationNumberForm
}

func PreviousApplicationNumber(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
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
				lpa.PreviousApplicationNumber = data.Form.PreviousApplicationNumber
				lpa.Tasks.YourDetails = actor.TaskCompleted

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
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
		validation.Empty())

	return errors
}
