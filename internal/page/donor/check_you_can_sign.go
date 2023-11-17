package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYouCanSignData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Options form.YesNoOptions
}

func CheckYouCanSign(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		data := &checkYouCanSignData{
			App: appData,
			Form: &form.YesNoForm{
				YesNo: lpa.Donor.CanSign,
			},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfYouWillBeAbleToSignYourself")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.Donor.CanSign = data.Form.YesNo

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				redirect := page.Paths.YourAddress
				if lpa.Donor.CanSign.IsNo() {
					redirect = page.Paths.NeedHelpSigningConfirmation
				}

				return redirect.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}
