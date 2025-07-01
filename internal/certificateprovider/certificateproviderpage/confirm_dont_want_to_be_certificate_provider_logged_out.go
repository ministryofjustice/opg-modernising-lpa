package certificateproviderpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderDataLoggedOut struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmDontWantToBeCertificateProviderLoggedOut(tmpl template.Template, accessCodeStore AccessCodeStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient, donorStore DonorStore, sessionStore SessionStore, notifyClient NotifyClient, donorStartURL string) page.Handler {
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
			code := accesscodedata.HashedFromQuery(r.URL.Query())

			accessCode, err := accessCodeStore.Get(r.Context(), actor.TypeCertificateProvider, code)
			if err != nil {
				return err
			}

			var email notify.Email

			if lpa.SignedForDonor() {
				email = notify.CertificateProviderOptedOutPostWitnessingEmail{
					Greeting:                      notifyClient.EmailGreeting(lpa),
					CertificateProviderFirstNames: lpa.CertificateProvider.FirstNames,
					CertificateProviderFullName:   lpa.CertificateProvider.FullName(),
					DonorFullName:                 lpa.Donor.FullName(),
					LpaType:                       appData.Localizer.T(lpa.Type.String()),
					LpaReferenceNumber:            lpa.LpaUID,
					DonorStartPageURL:             donorStartURL,
				}

				if !lpa.Status.IsCannotRegister() {
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
					LpaReferenceNumber:          donor.LpaUID,
					DonorStartPageURL:           donorStartURL,
				}

				donor.CertificateProvider = donordata.CertificateProvider{}
				donor.Tasks.CertificateProvider = task.StateNotStarted
				donor.Tasks.CheckYourLpa = task.StateNotStarted

				if err = donorStore.Put(ctx, donor); err != nil {
					return err
				}
			}

			if err := notifyClient.SendActorEmail(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, email); err != nil {
				return err
			}

			if err := accessCodeStore.Delete(r.Context(), accessCode); err != nil {
				return err
			}

			return page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider.RedirectQuery(w, r, appData, url.Values{
				"donorFullName":   {lpa.Donor.FullName()},
				"donorFirstNames": {lpa.Donor.FirstNames},
			})
		}

		return tmpl(w, data)
	}
}
