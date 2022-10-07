package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type readYourLpaData struct {
	App    AppData
	Errors map[string]string
	Lpa    Lpa
	Form   *readYourLpaForm
}

func ReadYourLpa(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &readYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &readYourLpaForm{
				Checked:   lpa.CheckedAgain,
				Confirm:   lpa.ConfirmFreeWill,
				Signature: lpa.SignatureCode,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readReadYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.CheckedAgain = data.Form.Checked
				lpa.ConfirmFreeWill = data.Form.Confirm
				lpa.SignatureCode = data.Form.Signature
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				appData.Lang.Redirect(w, r, signingConfirmationPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type readYourLpaForm struct {
	Checked   bool
	Confirm   bool
	Signature string
}

func readReadYourLpaForm(r *http.Request) *readYourLpaForm {
	r.ParseForm()

	return &readYourLpaForm{
		Checked:   postFormString(r, "checked") == "1",
		Confirm:   postFormString(r, "confirm") == "1",
		Signature: postFormString(r, "signature"),
	}
}

func (f *readYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !f.Checked {
		errors["checked"] = "selectReadAndCheckedLpa"
	}

	if !f.Confirm {
		errors["confirm"] = "selectConfirmMadeThisLpaOfOwnFreeWill"
	}

	if f.Signature == "" {
		errors["signature"] = "enterSignatureCode"
	}

	return errors
}
