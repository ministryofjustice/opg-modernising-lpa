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

func LpaType(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &lpaTypeData{
			App:  appData,
			Type: lpa.Type,
		}

		if r.Method == http.MethodPost {
			form := readLpaTypeForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.Type = form.LpaType
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.TaskList)
			}
		}

		return tmpl(w, data)
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

	if f.LpaType != LpaTypePropertyFinance && f.LpaType != LpaTypeHealthWelfare && f.LpaType != LpaTypeCombined {
		errors["lpa-type"] = "selectLpaType"
	}

	return errors
}
