package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type resendWitnessCodeData struct {
	App    page.AppData
	Errors validation.List
}

func ResendWitnessCode(tmpl template.Template, witnessCodeSender WitnessCodeSender, actorType actor.Type) Handler {
	send := witnessCodeSender.SendToCertificateProvider
	redirect := page.Paths.WitnessingAsCertificateProvider

	if actorType == actor.TypeIndependentWitness {
		send = witnessCodeSender.SendToIndependentWitness
		redirect = page.Paths.WitnessingAsIndependentWitness
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &resendWitnessCodeData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			if err := send(r.Context(), lpa, appData.Localizer); err != nil {
				if errors.Is(err, page.ErrTooManyWitnessCodeRequests) {
					data.Errors.Add("request", validation.CustomError{Label: "pleaseWaitOneMinute"})
					return tmpl(w, data)
				}

				return err
			}

			return redirect.Redirect(w, r, appData, lpa)
		}

		return tmpl(w, data)
	}
}
