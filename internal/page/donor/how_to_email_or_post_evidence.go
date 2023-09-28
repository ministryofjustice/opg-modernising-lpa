package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howToEmailOrPostEvidenceData struct {
	App    page.AppData
	Errors validation.List
}

func HowToEmailOrPostEvidence(tmpl template.Template, payer Payer) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howToEmailOrPostEvidenceData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			return payer.Pay(appData, w, r, lpa)
		}

		return tmpl(w, data)
	}
}
