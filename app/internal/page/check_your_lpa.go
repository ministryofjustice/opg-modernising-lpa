package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type checkYourLpaData struct {
	App       AppData
	Errors    map[string]string
	Lpa       *Lpa
	Form      *checkYourLpaForm
	Completed bool
}

func CheckYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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
			Completed: lpa.Tasks.CheckYourLpa == TaskCompleted,
		}

		if r.Method == http.MethodPost {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.Checked = data.Form.Checked
				lpa.HappyToShare = data.Form.Happy
				lpa.Tasks.CheckYourLpa = TaskCompleted

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.TaskList)
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

func (f *checkYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !f.Checked {
		errors["checked"] = "selectCheckedLpa"
	}

	if !f.Happy {
		errors["happy"] = "selectHappyToShareLpa"
	}

	return errors
}
