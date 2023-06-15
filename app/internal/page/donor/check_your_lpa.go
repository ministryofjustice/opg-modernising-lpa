package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type checkYourLpaData struct {
	App       page.AppData
	Errors    validation.List
	Lpa       *page.Lpa
	Form      *checkYourLpaForm
	Completed bool
}

func CheckYourLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &checkYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				Checked: lpa.Checked,
				Happy:   lpa.HappyToShare,
			},
			Completed: lpa.Tasks.CheckYourLpa.Completed(),
		}

		if r.Method == http.MethodPost {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.Checked = data.Form.Checked
				lpa.HappyToShare = data.Form.Happy
				lpa.Tasks.CheckYourLpa = actor.TaskCompleted

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.AboutPayment)
			}
		}

		return tmpl(w, data)
	}
}

type checkYourLpaForm struct {
	Checked bool
	Happy   bool
}

func readCheckYourLpaForm(r *http.Request) *checkYourLpaForm {
	return &checkYourLpaForm{
		Checked: page.PostFormString(r, "checked") == "1",
		Happy:   page.PostFormString(r, "happy") == "1",
	}
}

func (f *checkYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("checked", "checkedLpa", f.Checked,
		validation.Selected())
	errors.Bool("happy", "happyToShareLpa", f.Happy,
		validation.Selected())

	return errors
}
