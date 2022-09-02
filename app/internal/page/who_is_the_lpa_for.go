package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type whoIsTheLpaForData struct {
	Page             string
	L                localize.Localizer
	Lang             Lang
	CookieConsentSet bool
	Errors           map[string]string
	WhoFor           string
}

func WhoIsTheLpaFor(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, dataStore DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lpa Lpa
		dataStore.Get(r.Context(), sessionID(r), &lpa)

		data := &whoIsTheLpaForData{
			Page:             whoIsTheLpaForPath,
			L:                localizer,
			Lang:             lang,
			CookieConsentSet: cookieConsentSet(r),
			WhoFor:           lpa.WhoFor,
		}

		if r.Method == http.MethodPost {
			form := readWhoIsTheLpaForForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.WhoFor = form.WhoFor
				dataStore.Put(r.Context(), sessionID(r), lpa)
				lang.Redirect(w, r, donorDetailsPath, http.StatusFound)
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
