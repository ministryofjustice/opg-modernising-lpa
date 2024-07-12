package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type unableToConfirmIdentityData struct {
	App    page.AppData
	Donor  lpastore.Donor
	Errors validation.List
}

func UnableToConfirmIdentity(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			certificateProvider, err := certificateProviderStore.Get(r.Context())
			if err != nil {
				return err
			}

			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

			err = certificateProviderStore.Put(r.Context(), certificateProvider)
			if err != nil {
				return err
			}

			return page.Paths.CertificateProvider.ReadTheLpa.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &unableToConfirmIdentityData{
			App:   appData,
			Donor: lpa.Donor,
		})
	}
}
