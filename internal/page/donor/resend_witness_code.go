package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type resendWitnessCodeData struct {
	App    page.AppData
	Errors validation.List
}

func ResendWitnessCode(tmpl template.Template, witnessCodeSender WitnessCodeSender, now func() time.Time, actorType actor.Type) Handler {
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
			canRequest := lpa.CertificateProviderCodes.CanRequest

			if actorType == actor.TypeIndependentWitness {
				canRequest = lpa.IndependentWitnessCodes.CanRequest
			}

			if !canRequest(now()) {
				data.Errors.Add("request", validation.CustomError{Label: "pleaseWaitOneMinute"})
				return tmpl(w, data)
			}

			if err := send(r.Context(), lpa, appData.Localizer); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, redirect.Format(lpa.ID))
		}

		return tmpl(w, data)
	}
}
