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
	App       page.AppData
	Errors    validation.List
	Form      *form.LanguagePreferenceForm
	Options   localize.LangOptions
	FieldName string
}

func YourPreferredLanguage(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourPreferredLanguageData{
			App: appData,
			Form: &form.LanguagePreferenceForm{
				Preference: donor.ContactLanguagePreference,
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadLanguagePreferenceForm(r, "whichLanguageYoudLikeUsToUseWhenWeContactYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.ContactLanguagePreference = data.Form.Preference
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.LpaType.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
