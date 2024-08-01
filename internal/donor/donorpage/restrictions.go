package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type restrictionsData struct {
	App    page.AppData
	Errors validation.List
	Donor  *donordata.Provided
}

func Restrictions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &restrictionsData{
			App:   appData,
			Donor: donor,
		}

		if r.Method == http.MethodPost {
			form := readRestrictionsForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				donor.Tasks.Restrictions = task.StateCompleted
				donor.Restrictions = form.Restrictions

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
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
