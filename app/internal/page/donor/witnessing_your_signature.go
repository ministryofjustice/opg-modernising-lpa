package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type witnessingYourSignatureData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WitnessingYourSignature(tmpl template.Template, witnessCodeSender WitnessCodeSender) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			if err := witnessCodeSender.Send(r.Context(), lpa, appData.Localizer); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.WitnessingAsCertificateProvider)
		}

		data := &witnessingYourSignatureData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
