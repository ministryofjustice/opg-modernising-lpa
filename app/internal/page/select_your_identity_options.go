package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type selectYourIdentityOptionsData struct {
	App    AppData
	Errors map[string]string
	Form   *selectYourIdentityOptionsForm
	Page   int
}

func SelectYourIdentityOptions(tmpl template.Template, lpaStore LpaStore, page int) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &selectYourIdentityOptionsData{
			App:  appData,
			Page: page,
			Form: &selectYourIdentityOptionsForm{
				Selected: lpa.IdentityOption,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSelectYourIdentityOptionsForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.None {
				switch page {
				case 0:
					return appData.Lang.Redirect(w, r, lpa, Paths.SelectYourIdentityOptions1)
				case 1:
					return appData.Lang.Redirect(w, r, lpa, Paths.SelectYourIdentityOptions2)
				default:
					// will go to vouching flow when that is built
					return appData.Lang.Redirect(w, r, lpa, Paths.TaskList)
				}
			}

			if len(data.Errors) == 0 {
				lpa.IdentityOption = data.Form.Selected
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskInProgress

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.YourChosenIdentityOptions)
			}
		}

		return tmpl(w, data)
	}
}

type selectYourIdentityOptionsForm struct {
	Selected IdentityOption
	None     bool
}

func readSelectYourIdentityOptionsForm(r *http.Request) *selectYourIdentityOptionsForm {
	option := postFormString(r, "option")

	return &selectYourIdentityOptionsForm{
		Selected: readIdentityOption(option),
		None:     option == "none",
	}
}

func (f *selectYourIdentityOptionsForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.Selected == IdentityOptionUnknown && !f.None {
		errors["option"] = "selectAnIdentityOption"
	}

	return errors
}
