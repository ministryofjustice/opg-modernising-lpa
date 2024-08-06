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

type yourLpaLanguageData struct {
	App                appcontext.Data
	Errors             validation.List
	Form               *form.YesNoForm
	SelectedLanguage   localize.Lang
	UnselectedLanguage localize.Lang
}

func YourLpaLanguage(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourLpaLanguageData{
			App:              appData,
			Form:             form.NewYesNoForm(form.YesNoUnknown),
			SelectedLanguage: provided.Donor.LpaLanguagePreference,
		}

		if data.SelectedLanguage.IsEn() {
			data.UnselectedLanguage = localize.Cy
		} else {
			data.UnselectedLanguage = localize.En
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "whatYouWouldLikeToDo")
			data.Errors = f.Validate()

			if data.Errors.None() {
				if f.YesNo.IsNo() {
					provided.Donor.LpaLanguagePreference = data.UnselectedLanguage

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}
				}

				return donor.PathLpaYourLegalRightsAndResponsibilities.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
