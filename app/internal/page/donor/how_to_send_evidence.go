package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howToSendEvidenceData struct {
	App    page.AppData
	Errors validation.List
}

func HowToSendEvidence(tmpl template.Template, payer Payer) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howToSendEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			return payer.Pay(appData, w, r, lpa)
		}

		return tmpl(w, data)
	}
}
