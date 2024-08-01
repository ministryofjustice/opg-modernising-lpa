package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whenCanTheLpaBeUsedData struct {
	App     page.AppData
	Errors  validation.List
	Donor   *actor.DonorProvidedDetails
	Form    *whenCanTheLpaBeUsedForm
	Options donordata.CanBeUsedWhenOptions
}

func WhenCanTheLpaBeUsed(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &whenCanTheLpaBeUsedData{
			App:   appData,
			Donor: donor,
			Form: &whenCanTheLpaBeUsedForm{
				When: donor.WhenCanTheLpaBeUsed,
			},
			Options: donordata.CanBeUsedWhenValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhenCanTheLpaBeUsedForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.WhenCanTheLpaBeUsed = data.Form.When
				donor.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type whenCanTheLpaBeUsedForm struct {
	When  donordata.CanBeUsedWhen
	Error error
}

func readWhenCanTheLpaBeUsedForm(r *http.Request) *whenCanTheLpaBeUsedForm {
	when, err := donordata.ParseCanBeUsedWhen(page.PostFormString(r, "when"))

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
