package certificateprovider

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderDataLoggedOut struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ConfirmDontWantToBeCertificateProviderLoggedOut(tmpl template.Template, shareCodeStore ShareCodeStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient, donorStore DonorStore, sessionStore SessionStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		session, err := sessionStore.LpaData(r)
		if err != nil {
			return err
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: session.LpaID})

		lpa, err := lpaStoreResolvingService.Get(ctx)
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeCertificateProviderDataLoggedOut{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			if !lpa.SignedAt.IsZero() {
				if !lpa.CannotRegister {
					if err := lpaStoreClient.SendCertificateProviderOptOut(ctx, lpa.LpaUID, actoruid.Service); err != nil {
						return err
					}
				}
			} else {
				donor, err := donorStore.GetAny(ctx)
				if err != nil {
					return err
				}

				donor.CertificateProvider = actor.CertificateProvider{}
				donor.Tasks.CertificateProvider = actor.TaskNotStarted
				donor.Tasks.CheckYourLpa = actor.TaskNotStarted

				if err = donorStore.Put(ctx, donor); err != nil {
					return err
				}
			}

			shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, r.URL.Query().Get("referenceNumber"))
			if err != nil {
				return err
			}

			if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}
