package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoDoYouWantToBeCertificateProviderGuidanceData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WhoDoYouWantToBeCertificateProviderGuidance(tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whoDoYouWantToBeCertificateProviderGuidanceData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			if lpa.Tasks.CertificateProvider == actor.TaskNotStarted {
				lpa.Tasks.CertificateProvider = actor.TaskInProgress

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}
			}

			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderDetails)
		}

		return tmpl(w, data)
	}
}
