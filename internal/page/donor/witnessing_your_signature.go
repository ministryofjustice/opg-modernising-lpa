package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingYourSignatureData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WitnessingYourSignature(tmpl template.Template, witnessCodeSender WitnessCodeSender) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			if err := witnessCodeSender.SendToCertificateProvider(r.Context(), lpa, appData.Localizer); err != nil {
				return err
			}

			if lpa.Donor.CanSign.IsYes() {
				return appData.Redirect(w, r, lpa, page.Paths.WitnessingAsCertificateProvider.Format(lpa.ID))
			} else {
				if err := witnessCodeSender.SendToIndependentWitness(r.Context(), lpa, appData.Localizer); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.WitnessingAsIndependentWitness.Format(lpa.ID))
			}
		}

		data := &witnessingYourSignatureData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
