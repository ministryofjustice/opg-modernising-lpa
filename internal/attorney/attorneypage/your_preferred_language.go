package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App       appcontext.Data
	Errors    validation.List
	Form      *form.LanguagePreferenceForm
	Options   localize.LangOptions
	FieldName string
	Lpa       *lpadata.Lpa
}

func YourPreferredLanguage(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
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

				return attorney.PathConfirmYourDetails.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
