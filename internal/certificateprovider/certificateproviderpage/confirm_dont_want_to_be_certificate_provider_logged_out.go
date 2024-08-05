package certificateproviderpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderDataLoggedOut struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ConfirmDontWantToBeCertificateProviderLoggedOut(tmpl template.Template, shareCodeStore ShareCodeStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient, donorStore DonorStore, sessionStore SessionStore, notifyClient NotifyClient, appPublicURL string) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		session, err := sessionStore.LpaData(r)
		if err != nil {
			return err
		}

		ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: session.LpaID})

		lpa, err := lpaStoreResolvingService.Get(ctx)
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeCertificateProviderDataLoggedOut{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, r.URL.Query().Get("referenceNumber"))
			if err != nil {
				return err
			}

			var email notify.Email

			if !lpa.SignedAt.IsZero() {
				email = notify.CertificateProviderOptedOutPostWitnessingEmail{
					Greeting:                      notifyClient.EmailGreeting(lpa),
					CertificateProviderFirstNames: lpa.CertificateProvider.FirstNames,
					CertificateProviderFullName:   lpa.CertificateProvider.FullName(),
					DonorFullName:                 lpa.Donor.FullName(),
					LpaType:                       appData.Localizer.T(lpa.Type.String()),
					LpaUID:                        lpa.LpaUID,
					DonorStartPageURL:             appPublicURL + page.Paths.Start.Format(),
				}

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

				email = notify.CertificateProviderOptedOutPreWitnessingEmail{
					Greeting:                    notifyClient.EmailGreeting(lpa),
					CertificateProviderFullName: donor.CertificateProvider.FullName(),
					DonorFullName:               donor.Donor.FullName(),
					LpaType:                     appData.Localizer.T(donor.Type.String()),
					LpaUID:                      donor.LpaUID,
					DonorStartPageURL:           appPublicURL + page.Paths.Start.Format(),
				}

				donor.CertificateProvider = donordata.CertificateProvider{}
				donor.Tasks.CertificateProvider = task.StateNotStarted
				donor.Tasks.CheckYourLpa = task.StateNotStarted

				if err = donorStore.Put(ctx, donor); err != nil {
					return err
				}
			}

			if err := notifyClient.SendActorEmail(ctx, lpa.CorrespondentEmail(), lpa.LpaUID, email); err != nil {
				return err
			}

			if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.YouHaveDecidedNotToBeCertificateProvider.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}
