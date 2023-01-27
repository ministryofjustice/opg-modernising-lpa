package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourLpaData struct {
	App       AppData
	Errors    validation.List
	Lpa       *Lpa
	Form      *checkYourLpaForm
	Completed bool
}

func CheckYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

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

			if data.Errors.Empty() {
				lpa.Checked = data.Form.Checked
				lpa.HappyToShare = data.Form.Happy
				lpa.Tasks.CheckYourLpa = TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.TaskList)
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
	r.ParseForm()

	return &checkYourLpaForm{
		Checked: postFormString(r, "checked") == "1",
		Happy:   postFormString(r, "happy") == "1",
	}
}

func (f *checkYourLpaForm) Validate() validation.List {
	var errors validation.List

	if !f.Checked {
		errors.Add("checked", "selectCheckedLpa")
	}

	if !f.Happy {
		errors.Add("happy", "selectHappyToShareLpa")
	}

	return errors
}
