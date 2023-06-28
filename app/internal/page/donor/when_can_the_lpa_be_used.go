package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type whenCanTheLpaBeUsedOptions struct {
	WhenCapacityLost page.CanBeUsedWhen
	WhenRegistered   page.CanBeUsedWhen
}

type whenCanTheLpaBeUsedData struct {
	App     page.AppData
	Errors  validation.List
	Lpa     *page.Lpa
	Form    *whenCanTheLpaBeUsedForm
	Options whenCanTheLpaBeUsedOptions
}

func WhenCanTheLpaBeUsed(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &whenCanTheLpaBeUsedData{
			App: appData,
			Lpa: lpa,
			Form: &whenCanTheLpaBeUsedForm{
				When: lpa.WhenCanTheLpaBeUsed,
			},
			Options: whenCanTheLpaBeUsedOptions{
				WhenCapacityLost: page.CanBeUsedWhenCapacityLost,
				WhenRegistered:   page.CanBeUsedWhenRegistered,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readWhenCanTheLpaBeUsedForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.WhenCanTheLpaBeUsed = data.Form.When
				lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type whenCanTheLpaBeUsedForm struct {
	When  page.CanBeUsedWhen
	Error error
}

func readWhenCanTheLpaBeUsedForm(r *http.Request) *whenCanTheLpaBeUsedForm {
	when, err := page.ParseCanBeUsedWhen(page.PostFormString(r, "when"))

	return &whenCanTheLpaBeUsedForm{
		When:  when,
		Error: err,
	}
}

func (f *whenCanTheLpaBeUsedForm) Validate() validation.List {
	var errors validation.List

	errors.Error("when", "whenYourAttorneysCanUseYourLpa", f.Error,
		validation.Selected())

	return errors
}
