package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider *actor.CertificateProviderProvidedDetails
}

func OneLoginIdentityDetails(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &oneLoginIdentityDetailsData{
			App:                 appData,
			CertificateProvider: certificateProvider,
		}

		if r.Method == http.MethodPost {
			lpa, err := lpaStoreResolvingService.Get(r.Context())
			if err != nil {
				return err
			}

			if certificateProvider.CertificateProviderIdentityConfirmed(
				lpa.CertificateProvider.FirstNames,
				lpa.CertificateProvider.LastName,
			) {
				certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

				if err = certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return page.Paths.CertificateProvider.ReadTheLpa.Redirect(w, r, appData, certificateProvider.LpaID)
			} else {
				// TODO: will be changed in MLPAB-2234
				return page.Paths.CertificateProvider.ProveYourIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
