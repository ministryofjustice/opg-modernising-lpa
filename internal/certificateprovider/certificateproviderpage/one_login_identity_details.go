package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider *certificateproviderdata.Provided
}

func OneLoginIdentityDetails(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
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
				certificateProvider.Tasks.ConfirmYourIdentity = task.StateCompleted

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
