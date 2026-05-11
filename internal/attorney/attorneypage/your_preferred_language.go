package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *newforms.LanguageForm
	Lpa    *lpadata.Lpa
}

func YourPreferredLanguage(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
		data := &yourPreferredLanguageData{
			App:  appData,
			Form: newforms.NewLanguageForm(appData.Localizer.T("whichLanguageYouWouldLikeUsToUseWhenWeContactYou")),
			Lpa:  lpa,
		}

		data.Form.Language.Input = attorneyProvidedDetails.ContactLanguagePreference.String()

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				attorneyProvidedDetails.ContactLanguagePreference = data.Form.Language.Value
				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				return attorney.PathConfirmYourDetails.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
