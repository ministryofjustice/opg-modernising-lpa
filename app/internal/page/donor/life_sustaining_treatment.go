package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type lifeSustainingTreatmentData struct {
	App     page.AppData
	Errors  validation.List
	Form    *lifeSustainingTreatmentForm
	Options page.LifeSustainingTreatmentOptions
}

func LifeSustainingTreatment(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &lifeSustainingTreatmentData{
			App: appData,
			Form: &lifeSustainingTreatmentForm{
				Option: lpa.LifeSustainingTreatmentOption,
			},
			Options: page.LifeSustainingTreatmentValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readLifeSustainingTreatmentForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.LifeSustainingTreatmentOption = data.Form.Option
				lpa.Tasks.LifeSustainingTreatment = actor.TaskCompleted
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type lifeSustainingTreatmentForm struct {
	Option page.LifeSustainingTreatment
	Error  error
}

func readLifeSustainingTreatmentForm(r *http.Request) *lifeSustainingTreatmentForm {
	option, err := page.ParseLifeSustainingTreatment(page.PostFormString(r, "option"))

	return &lifeSustainingTreatmentForm{
		Option: option,
		Error:  err,
	}
}

func (f *lifeSustainingTreatmentForm) Validate() validation.List {
	var errors validation.List

	errors.Error("option", "ifTheDonorGivesConsentToLifeSustainingTreatment", f.Error,
		validation.Selected())

	return errors
}
