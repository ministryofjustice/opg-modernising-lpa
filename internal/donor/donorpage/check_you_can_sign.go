package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYouCanSignData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
}

func CheckYouCanSign(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &checkYouCanSignData{
			App:  appData,
			Form: form.NewYesNoForm(provided.Donor.CanSign),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfYouWillBeAbleToSignYourself")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.CanSign = data.Form.YesNo

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				redirect := donor.PathYourPreferredLanguage
				if provided.Donor.CanSign.IsNo() {
					redirect = donor.PathNeedHelpSigningConfirmation
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
