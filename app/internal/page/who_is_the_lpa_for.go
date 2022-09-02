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

func WhoIsTheLpaFor(logger Logger, tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) {
		var lpa Lpa
		dataStore.Get(r.Context(), appData.SessionID, &lpa)

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
				return
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
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
