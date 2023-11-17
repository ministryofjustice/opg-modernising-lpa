package donor


import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type restrictionsData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *actor.Lpa
}

func Restrictions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
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

				return page.Paths.TaskList.Redirect(w, r, appData, lpa)
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
