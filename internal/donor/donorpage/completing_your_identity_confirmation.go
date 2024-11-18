package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type completingYourIdentityConfirmationData struct {
	App      appcontext.Data
	Errors   validation.List
	Form     *howWillYouConfirmYourIdentityForm
	Options  howYouWillConfirmYourIdentityOptions
	Deadline time.Time
}

func CompletingYourIdentityConfirmation(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &completingYourIdentityConfirmationData{
			App:      appData,
			Form:     &howWillYouConfirmYourIdentityForm{},
			Options:  howYouWillConfirmYourIdentityValues,
			Deadline: provided.IdentityDeadline(),
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWillYouConfirmYourIdentityForm(r, "howYouWouldLikeToContinue")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.How {
				case howYouWillConfirmYourIdentityWithdraw:
					if provided.WitnessedByCertificateProviderAt.IsZero() {
						return donor.PathDeleteThisLpa.Redirect(w, r, appData, provided)
					}

					return donor.PathWithdrawThisLpa.Redirect(w, r, appData, provided)

				default:
					return donor.PathIdentityWithOneLogin.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
