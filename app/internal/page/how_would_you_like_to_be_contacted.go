package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type howWouldYouLikeToBeContactedData struct {
	Page             string
	L                localize.Localizer
	Lang             Lang
	CookieConsentSet bool
	Errors           map[string]string
	Contact          []string
}

func HowWouldYouLikeToBeContacted(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, dataStore DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lpa Lpa
		dataStore.Get(r.Context(), sessionID(r), &lpa)

		data := &howWouldYouLikeToBeContactedData{
			Page:             howWouldYouLikeToBeContactedPath,
			L:                localizer,
			Lang:             lang,
			CookieConsentSet: cookieConsentSet(r),
			Contact:          lpa.Contact,
		}

		if r.Method == http.MethodPost {
			form := readHowWouldYouLikeToBeContactedForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.Contact = form.Contact
				dataStore.Put(r.Context(), sessionID(r), lpa)
				lang.Redirect(w, r, taskListPath, http.StatusFound)
				return
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}

type howWouldYouLikeToBeContactedForm struct {
	Contact []string
}

func readHowWouldYouLikeToBeContactedForm(r *http.Request) *howWouldYouLikeToBeContactedForm {
	r.ParseForm()

	return &howWouldYouLikeToBeContactedForm{
		Contact: r.PostForm["contact"],
	}
}

func (f *howWouldYouLikeToBeContactedForm) Validate() map[string]string {
	errors := map[string]string{}

	if len(f.Contact) == 0 {
		errors["contact"] = "selectContact"
	}

	for _, value := range f.Contact {
		if value != "email" && value != "phone" && value != "text message" && value != "post" {
			errors["contact"] = "selectContact"
			break
		}
	}

	return errors
}
