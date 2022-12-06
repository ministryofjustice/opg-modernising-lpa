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

func SelectYourIdentityOptions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &selectYourIdentityOptionsData{
			App: appData,
			Form: &selectYourIdentityOptionsForm{
				Options: lpa.IdentityOptions.Selected,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSelectYourIdentityOptionsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.IdentityOptions = IdentityOptions{
					Selected: data.Form.Options,
					First:    data.Form.First,
					Second:   data.Form.Second,
				}
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskInProgress

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, appData.Paths.YourChosenIdentityOptions, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type selectYourIdentityOptionsForm struct {
	Options       []IdentityOption
	First, Second IdentityOption
}

func readSelectYourIdentityOptionsForm(r *http.Request) *selectYourIdentityOptionsForm {
	r.ParseForm()

	mappedOptions := make([]IdentityOption, len(r.PostForm["options"]))
	for i, option := range r.PostForm["options"] {
		mappedOptions[i] = readIdentityOption(option)
	}

	first, second := identityOptionsRanked(mappedOptions)

	return &selectYourIdentityOptionsForm{
		Options: mappedOptions,
		First:   first,
		Second:  second,
	}
}

func (f *selectYourIdentityOptionsForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.First == IdentityOptionUnknown || f.Second == IdentityOptionUnknown {
		errors["options"] = "selectMoreOptions"
	}

	if len(f.Options) < 3 {
		errors["options"] = "selectAtLeastThreeIdentityOptions"
	}

	if slices.Contains(f.Options, IdentityOptionUnknown) {
		errors["options"] = "selectValidIdentityOption"
	}

	return errors
}
