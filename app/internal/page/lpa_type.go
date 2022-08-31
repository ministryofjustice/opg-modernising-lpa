package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type lpaTypeData struct {
	Page   string
	L      localize.Localizer
	Lang   Lang
	Errors map[string]string
}

func LpaType(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, dataStore DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &lpaTypeData{
			Page: lpaTypePath,
			L:    localizer,
			Lang: lang,
		}

		if r.Method == http.MethodPost {
			form := readLpaTypeForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				dataStore.Save(form.LpaType)
				lang.Redirect(w, r, whoIsTheLpaForPath, http.StatusFound)
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
