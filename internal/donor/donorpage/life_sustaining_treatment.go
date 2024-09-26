package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lifeSustainingTreatmentData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *lifeSustainingTreatmentForm
	Options lpadata.LifeSustainingTreatmentOptions
}

func LifeSustainingTreatment(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &lifeSustainingTreatmentData{
			App: appData,
			Form: &lifeSustainingTreatmentForm{
				Option: provided.LifeSustainingTreatmentOption,
			},
			Options: lpadata.LifeSustainingTreatmentValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readLifeSustainingTreatmentForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.LifeSustainingTreatmentOption = data.Form.Option
				provided.Tasks.LifeSustainingTreatment = task.StateCompleted
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type lifeSustainingTreatmentForm struct {
	Option lpadata.LifeSustainingTreatment
}

func readLifeSustainingTreatmentForm(r *http.Request) *lifeSustainingTreatmentForm {
	option, _ := lpadata.ParseLifeSustainingTreatment(page.PostFormString(r, "option"))

	return &lifeSustainingTreatmentForm{
		Option: option,
	}
}

func (f *lifeSustainingTreatmentForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("option", "ifTheDonorGivesConsentToLifeSustainingTreatment", f.Option,
		validation.Selected())

	return errors
}
