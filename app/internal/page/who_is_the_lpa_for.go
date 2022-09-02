package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type whoIsTheLpaForData struct {
	App    AppData
	Errors map[string]string
	WhoFor string
}

func WhoIsTheLpaFor(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &whoIsTheLpaForData{
			App:    appData,
			WhoFor: lpa.WhoFor,
		}

		if r.Method == http.MethodPost {
			form := readWhoIsTheLpaForForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.WhoFor = form.WhoFor
				dataStore.Put(r.Context(), appData.SessionID, lpa)
				appData.Lang.Redirect(w, r, donorDetailsPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type whoIsTheLpaForForm struct {
	WhoFor string
}

func readWhoIsTheLpaForForm(r *http.Request) *whoIsTheLpaForForm {
	return &whoIsTheLpaForForm{
		WhoFor: postFormString(r, "who-for"),
	}
}

func (f *whoIsTheLpaForForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.WhoFor != "me" && f.WhoFor != "someone-else" {
		errors["who-for"] = "selectWhoFor"
	}

	return errors
}
