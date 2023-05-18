package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type resendWitnessCodeData struct {
	App    page.AppData
	Errors validation.List
}

func ResendWitnessCode(tmpl template.Template, donorStore DonorStore, witnessCodeSender WitnessCodeSender, now func() time.Time) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &resendWitnessCodeData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			if !lpa.WitnessCodes.CanRequest(now()) {
				data.Errors.Add("request", validation.CustomError{Label: "pleaseWaitOneMinute"})
				return tmpl(w, data)
			}

			if err := witnessCodeSender.Send(r.Context(), lpa, appData.Localizer); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.WitnessingAsCertificateProvider)
		}

		return tmpl(w, data)
	}
}
