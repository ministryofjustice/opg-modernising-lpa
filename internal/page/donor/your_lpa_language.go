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

type yourLpaLanguageData struct {
	App                page.AppData
	Errors             validation.List
	Form               *form.YesNoForm
	SelectedLanguage   localize.Lang
	UnselectedLanguage localize.Lang
}

func YourLpaLanguage(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourLpaLanguageData{
			App:              appData,
			Form:             form.NewYesNoForm(form.YesNoUnknown),
			SelectedLanguage: donor.Donor.LpaLanguagePreference,
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
					donor.Donor.LpaLanguagePreference = data.UnselectedLanguage

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				return page.Paths.LpaYourLegalRightsAndResponsibilities.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
