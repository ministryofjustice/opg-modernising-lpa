package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *yourPreferredLanguageForm
	Options     localize.LangOptions
	CanTaskList bool
}

func YourPreferredLanguage(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourPreferredLanguageData{
			App: appData,
			Form: &yourPreferredLanguageForm{
				Contact: provided.Donor.ContactLanguagePreference,
				Lpa:     provided.Donor.LpaLanguagePreference,
			},
			Options:     localize.LangValues,
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourPreferredLanguageForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.ContactLanguagePreference = data.Form.Contact
				provided.Donor.LpaLanguagePreference = data.Form.Lpa

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathYourLegalRightsAndResponsibilitiesIfYouMakeLpa.Redirect(w, r, appData, provided)
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
