package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lifeSustainingTreatmentData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *lifeSustainingTreatmentForm
	Options donordata.LifeSustainingTreatmentOptions
}

func LifeSustainingTreatment(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &lifeSustainingTreatmentData{
			App: appData,
			Form: &lifeSustainingTreatmentForm{
				Option: donor.LifeSustainingTreatmentOption,
			},
			Options: donordata.LifeSustainingTreatmentValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readLifeSustainingTreatmentForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.LifeSustainingTreatmentOption = data.Form.Option
				donor.Tasks.LifeSustainingTreatment = task.StateCompleted
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type lifeSustainingTreatmentForm struct {
	Option donordata.LifeSustainingTreatment
	Error  error
}

func readLifeSustainingTreatmentForm(r *http.Request) *lifeSustainingTreatmentForm {
	option, err := donordata.ParseLifeSustainingTreatment(page.PostFormString(r, "option"))

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
