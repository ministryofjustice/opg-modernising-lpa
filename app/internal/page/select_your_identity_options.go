package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"golang.org/x/exp/slices"
)

type selectYourIdentityOptionsData struct {
	App    AppData
	Errors map[string]string
	Form   *selectYourIdentityOptionsForm
}

func SelectYourIdentityOptions(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &selectYourIdentityOptionsData{
			App: appData,
			Form: &selectYourIdentityOptionsForm{
				Options: lpa.IdentityOptions,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSelectYourIdentityOptionsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.IdentityOptions = data.Form.Options

				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, yourChosenIdentityOptionsPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type selectYourIdentityOptionsForm struct {
	Options []IdentityOption
}

func readSelectYourIdentityOptionsForm(r *http.Request) *selectYourIdentityOptionsForm {
	r.ParseForm()

	mappedOptions := make([]IdentityOption, len(r.PostForm["options"]))
	for i, option := range r.PostForm["options"] {
		mappedOptions[i] = readIdentityOption(option)
	}

	return &selectYourIdentityOptionsForm{
		Options: mappedOptions,
	}
}

func (f *selectYourIdentityOptionsForm) Validate() map[string]string {
	errors := map[string]string{}

	if len(f.Options) < 3 {
		errors["options"] = "selectAtLeastThreeIdentityOptions"
	}

	if slices.Contains(f.Options, IdentityOptionUnknown) {
		errors["options"] = "selectValidIdentityOption"
	}

	return errors
}
