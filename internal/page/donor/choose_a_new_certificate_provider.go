package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseNewCertificateProviderData struct {
	Lpa    *page.Lpa
	Errors validation.List
	App    page.AppData
}

func ChooseNewCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &chooseNewCertificateProviderData{Lpa: lpa, App: appData}

		if r.Method == http.MethodPost {
			lpa.CertificateProvider = actor.CertificateProvider{}

			if err := donorStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.ChooseYourCertificateProvider.Format(lpa.ID))
		}

		return tmpl(w, data)
	}
}
