package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoDoYouWantToBeCertificateProviderGuidanceData struct {
	App        page.AppData
	Errors     validation.List
	NotStarted bool
	Lpa        *page.Lpa
}

func WhoDoYouWantToBeCertificateProviderGuidance(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whoDoYouWantToBeCertificateProviderGuidanceData{
			App:        appData,
			NotStarted: lpa.Tasks.CertificateProvider == page.TaskNotStarted,
			Lpa:        lpa,
		}

		if r.Method == http.MethodPost {
			if page.PostFormString(r, "will-do-this-later") == "1" {
				return appData.Redirect(w, r, lpa, page.Paths.TaskList)
			}

			if data.NotStarted {
				lpa.Tasks.CertificateProvider = page.TaskInProgress
			}
			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderDetails)
		}

		return tmpl(w, data)
	}
}
