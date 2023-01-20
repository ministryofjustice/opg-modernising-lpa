package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type whoDoYouWantToBeCertificateProviderGuidanceData struct {
	App        AppData
	Errors     map[string]string
	NotStarted bool
	Lpa        *Lpa
}

func WhoDoYouWantToBeCertificateProviderGuidance(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whoDoYouWantToBeCertificateProviderGuidanceData{
			App:        appData,
			NotStarted: lpa.Tasks.CertificateProvider == TaskNotStarted,
			Lpa:        lpa,
		}

		if r.Method == http.MethodPost {
			if postFormString(r, "will-do-this-later") == "1" {
				return appData.Redirect(w, r, lpa, Paths.TaskList)
			}

			if data.NotStarted {
				lpa.Tasks.CertificateProvider = TaskInProgress
			}
			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, Paths.CertificateProviderDetails)
		}

		return tmpl(w, data)
	}
}
