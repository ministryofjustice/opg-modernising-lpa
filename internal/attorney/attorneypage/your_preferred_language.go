package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App       page.AppData
	Errors    validation.List
	Form      *form.LanguagePreferenceForm
	Options   localize.LangOptions
	FieldName string
	Lpa       *lpastore.Lpa
}

func YourPreferredLanguage(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourPreferredLanguageData{
			App: appData,
			Form: &form.LanguagePreferenceForm{
				Preference: attorneyProvidedDetails.ContactLanguagePreference,
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Lpa:       lpa,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadLanguagePreferenceForm(r, "whichLanguageYouWouldLikeUsToUseWhenWeContactYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.ContactLanguagePreference = data.Form.Preference
				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				return page.Paths.Attorney.ConfirmYourDetails.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
