package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App     page.AppData
	Errors  validation.List
	Form    *yourPreferredLanguageForm
	Options localize.LangOptions
}

func YourPreferredLanguage(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourPreferredLanguageData{
			App: appData,
			Form: &yourPreferredLanguageForm{
				Contact: donor.Donor.ContactLanguagePreference,
				Lpa:     donor.Donor.LpaLanguagePreference,
			},
			Options: localize.LangValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readYourPreferredLanguageForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Donor.ContactLanguagePreference = data.Form.Contact
				donor.Donor.LpaLanguagePreference = data.Form.Lpa

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.YourLegalRightsAndResponsibilitiesIfYouMakeLpa.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type yourPreferredLanguageForm struct {
	Contact      localize.Lang
	ContactError error
	Lpa          localize.Lang
	LpaError     error
}

func readYourPreferredLanguageForm(r *http.Request) *yourPreferredLanguageForm {
	contact, contactErr := localize.ParseLang(form.PostFormString(r, "contact-language"))
	lpa, lpaErr := localize.ParseLang(form.PostFormString(r, "lpa-language"))

	return &yourPreferredLanguageForm{
		Contact:      contact,
		ContactError: contactErr,
		Lpa:          lpa,
		LpaError:     lpaErr,
	}
}

func (f *yourPreferredLanguageForm) Validate() validation.List {
	var errors validation.List

	errors.Error("contact-language", "whichLanguageYouWouldLikeUsToUseWhenWeContactYou", f.ContactError,
		validation.Selected())
	errors.Error("lpa-language", "theLanguageInWhichYouWouldLikeYourLpaRegistered", f.LpaError,
		validation.Selected())

	return errors
}
