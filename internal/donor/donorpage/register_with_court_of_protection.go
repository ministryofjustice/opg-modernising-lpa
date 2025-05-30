package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type registerWithCourtOfProtectionData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func RegisterWithCourtOfProtection(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &registerWithCourtOfProtectionData{
			App:   appData,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
			Donor: provided,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "whatYouWouldLikeToDo")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo.IsYes() {
					return donor.PathDeleteThisLpa.Redirect(w, r, appData, provided)
				} else {
					provided.RegisteringWithCourtOfProtection = true

					if err := eventClient.SendRegisterWithCourtOfProtection(r.Context(), event.RegisterWithCourtOfProtection{
						UID: provided.LpaUID,
					}); err != nil {
						return err
					}
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
