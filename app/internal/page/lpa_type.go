package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type lpaTypeData struct {
	App    AppData
	Errors map[string]string
	Type   string
}

func LpaType(logger Logger, tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) {
		var lpa Lpa
		dataStore.Get(r.Context(), appData.SessionID, &lpa)

		data := &lpaTypeData{
			App:  appData,
			Type: lpa.Type,
		}

		if r.Method == http.MethodPost {
			form := readLpaTypeForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.Type = form.LpaType
				dataStore.Put(r.Context(), appData.SessionID, lpa)
				appData.Lang.Redirect(w, r, whoIsTheLpaForPath, http.StatusFound)
				return
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}

type lpaTypeForm struct {
	LpaType string
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	return &lpaTypeForm{
		LpaType: postFormString(r, "lpa-type"),
	}
}

func (f *lpaTypeForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.LpaType != "pfa" && f.LpaType != "hw" && f.LpaType != "both" {
		errors["lpa-type"] = "selectLpaType"
	}

	return errors
}
