package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                 page.AppData
	Errors              validation.List
	Donor               *actor.DonorProvidedDetails
	CertificateProvider *actor.CertificateProviderProvidedDetails
}

func ConfirmYourDetails(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		donor, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			redirect := page.Paths.CertificateProvider.YourRole
			if donor.Tasks.ConfirmYourIdentityAndSign.Completed() {
				redirect = page.Paths.CertificateProvider.TaskList
			}

			return redirect.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		data := &confirmYourDetailsData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Donor:               donor,
		}

		return tmpl(w, data)
	}
}
