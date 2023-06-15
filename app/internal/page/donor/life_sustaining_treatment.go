package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type lifeSustainingTreatmentData struct {
	App    page.AppData
	Errors validation.List
	Option string
}

func LifeSustainingTreatment(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &lifeSustainingTreatmentData{
			App:    appData,
			Option: lpa.LifeSustainingTreatmentOption,
		}

		if r.Method == http.MethodPost {
			form := readLifeSustainingTreatmentForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.LifeSustainingTreatmentOption = form.Option
				lpa.Tasks.LifeSustainingTreatment = actor.TaskCompleted
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}

type lifeSustainingTreatmentForm struct {
	Option string
}

func readLifeSustainingTreatmentForm(r *http.Request) *lifeSustainingTreatmentForm {
	return &lifeSustainingTreatmentForm{
		Option: page.PostFormString(r, "option"),
	}
}

func (f *lifeSustainingTreatmentForm) Validate() validation.List {
	var errors validation.List

	errors.String("option", "ifTheDonorGivesConsentToLifeSustainingTreatment", f.Option,
		validation.Select(page.OptionA, page.OptionB))

	return errors
}
