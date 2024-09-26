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

type whenCanTheLpaBeUsedData struct {
	App     appcontext.Data
	Errors  validation.List
	Donor   *donordata.Provided
	Form    *whenCanTheLpaBeUsedForm
	Options lpadata.CanBeUsedWhenOptions
}

func WhenCanTheLpaBeUsed(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whenCanTheLpaBeUsedData{
			App:   appData,
			Donor: provided,
			Form: &whenCanTheLpaBeUsedForm{
				When: provided.WhenCanTheLpaBeUsed,
			},
			Options: lpadata.CanBeUsedWhenValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhenCanTheLpaBeUsedForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.WhenCanTheLpaBeUsed = data.Form.When
				provided.Tasks.WhenCanTheLpaBeUsed = task.StateCompleted
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type whenCanTheLpaBeUsedForm struct {
	When lpadata.CanBeUsedWhen
}

func readWhenCanTheLpaBeUsedForm(r *http.Request) *whenCanTheLpaBeUsedForm {
	when, _ := lpadata.ParseCanBeUsedWhen(page.PostFormString(r, "when"))

	return &whenCanTheLpaBeUsedForm{
		When: when,
	}
}

func (f *whenCanTheLpaBeUsedForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("when", "whenYourAttorneysCanUseYourLpa", f.When,
		validation.Selected())

	return errors
}
