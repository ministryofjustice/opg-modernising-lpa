package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App           appcontext.Data
	Errors        validation.List
	Provided      *certificateproviderdata.Provided
	DonorFullName string
}

func OneLoginIdentityDetails(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &oneLoginIdentityDetailsData{
			App:           appData,
			Provided:      certificateProvider,
			DonorFullName: lpa.Donor.FullName(),
		}

		if r.Method == http.MethodPost {
			if certificateProvider.CertificateProviderIdentityConfirmed(
				lpa.CertificateProvider.FirstNames,
				lpa.CertificateProvider.LastName,
			) {
				certificateProvider.Tasks.ConfirmYourIdentity = task.StateCompleted

				if err = certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return certificateprovider.PathReadTheLpa.Redirect(w, r, appData, certificateProvider.LpaID)
			} else {
				// TODO: will be changed in MLPAB-2234
				return certificateprovider.PathProveYourIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
