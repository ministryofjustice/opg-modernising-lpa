package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseNewCertificateProviderData struct {
	Donor  *actor.DonorProvidedDetails
	Errors validation.List
	App    page.AppData
}

func ChooseNewCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &chooseNewCertificateProviderData{Donor: donor, App: appData}

		if r.Method == http.MethodPost {
			donor.CertificateProvider = donordata.CertificateProvider{}

			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			return page.Paths.ChooseYourCertificateProvider.Redirect(w, r, appData, donor)
		}

		return tmpl(w, data)
	}
}
