package certificateprovider

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ConfirmDontWantToBeCertificateProvider(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeCertificateProviderData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			if !lpa.SignedAt.IsZero() {
				if !lpa.CannotRegister {
					if err := lpaStoreClient.SendCertificateProviderOptOut(r.Context(), lpa.LpaUID, lpa.CertificateProvider.UID); err != nil {
						return err
					}
				}
			} else {
				donor, err := donorStore.GetAny(r.Context())
				if err != nil {
					return err
				}

				donor.CertificateProvider = actor.CertificateProvider{}
				donor.Tasks.CertificateProvider = actor.TaskNotStarted
				donor.Tasks.CheckYourLpa = actor.TaskNotStarted

				if err = donorStore.Put(r.Context(), donor); err != nil {
					return err
				}
			}

			if err := certificateProviderStore.Delete(r.Context()); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}
