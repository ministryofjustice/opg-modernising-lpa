package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whenCanTheLpaBeUsedData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
	Form   *form.SelectForm[lpadata.CanBeUsedWhen, lpadata.CanBeUsedWhenOptions, *lpadata.CanBeUsedWhen]
}

func WhenCanTheLpaBeUsed(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whenCanTheLpaBeUsedData{
			App:   appData,
			Donor: provided,
			Form:  form.NewSelectForm(provided.WhenCanTheLpaBeUsed, lpadata.CanBeUsedWhenValues, "whenYourAttorneysCanUseYourLpa"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.WhenCanTheLpaBeUsed = data.Form.Selected
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
