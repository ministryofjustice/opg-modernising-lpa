package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type restrictionsData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func Restrictions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &restrictionsData{
			App:   appData,
			Donor: provided,
		}

		if r.Method == http.MethodPost {
			form := readRestrictionsForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				provided.Tasks.Restrictions = task.StateCompleted
				provided.Restrictions = form.Restrictions

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type restrictionsForm struct {
	Restrictions string
}

func readRestrictionsForm(r *http.Request) *restrictionsForm {
	return &restrictionsForm{
		Restrictions: page.PostFormString(r, "restrictions"),
	}
}

func (f *restrictionsForm) Validate() validation.List {
	var errors validation.List

	errors.String("restrictions", "restrictions", f.Restrictions,
		validation.StringTooLong(10000))

	return errors
}
