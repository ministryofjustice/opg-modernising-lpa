package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type restrictionsData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func Restrictions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &restrictionsData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			form := readRestrictionsForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.Tasks.Restrictions = actor.TaskCompleted
				lpa.Restrictions = form.Restrictions

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
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
