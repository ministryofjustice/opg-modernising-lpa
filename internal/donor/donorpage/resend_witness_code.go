package donorpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type resendWitnessCodeData struct {
	App    appcontext.Data
	Errors validation.List
}

func ResendWitnessCode(tmpl template.Template, witnessCodeSender WitnessCodeSender, actorType actor.Type) Handler {
	send := witnessCodeSender.SendToCertificateProvider
	redirect := donor.PathWitnessingAsCertificateProvider

	if actorType == actor.TypeIndependentWitness {
		send = witnessCodeSender.SendToIndependentWitness
		redirect = donor.PathWitnessingAsIndependentWitness
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &resendWitnessCodeData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			if err := send(r.Context(), provided); err != nil {
				if errors.Is(err, donor.ErrTooManyWitnessCodeRequests) {
					data.Errors.Add("request", validation.CustomError{Label: "pleaseWaitOneMinute"})
					return tmpl(w, data)
				}

				return err
			}

			return redirect.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
