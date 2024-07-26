package certificateproviderpage

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

func UnableToConfirmIdentity(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, certificateProvider *actor.CertificateProviderProvidedDetails) error {
		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
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
