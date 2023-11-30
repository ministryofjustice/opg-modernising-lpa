package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App        page.AppData
	Errors     validation.List
	Form       *form.LanguagePreferenceForm
	Options    localize.LangOptions
	FieldNames form.Names
}

func YourPreferredLanguage(tmpl template.Template, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourPreferredLanguageData{
			App: appData,
			Form: &form.LanguagePreferenceForm{
				Preference: certificateProvider.ContactLanguagePreference,
			},
			Options:    localize.LangValues,
			FieldNames: form.FieldNames,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadLanguagePreferenceForm(r, "yourPreferredLanguageForWhenWeContactYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				certificateProvider.ContactLanguagePreference = data.Form.Preference
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return page.Paths.CertificateProvider.ConfirmYourDetails.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
