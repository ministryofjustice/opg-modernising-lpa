package certificateprovider

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService, notifyClient NotifyClient, lpaStoreClient LpaStoreClient) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			return page.Paths.CertificateProvider.OneLoginIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		if r.FormValue("error") == "access_denied" {
			// TODO: check with team on how we want to communicate this on the page
			return errors.New("access denied")
		}

		oneLoginSession, err := sessionStore.OneLogin(r)
		if err != nil {
			return err
		}

		_, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		userData, err := oneLoginClient.ParseIdentityClaim(r.Context(), userInfo)
		if err != nil {
			return err
		}

		certificateProvider.IdentityUserData = userData

		if err = certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
			return err
		}

		switch certificateProvider.IdentityUserData.Status {
		case identity.StatusFailed, identity.StatusInsufficientEvidence, identity.StatusUnknown:
			if !lpa.SignedAt.IsZero() {
				if err = notifyClient.SendActorEmail(r.Context(), lpa.Donor.Email, lpa.LpaUID, notify.CertificateProviderFailedIDCheckEmail{
					DonorFullName:               lpa.Donor.FullName(),
					CertificateProviderFullName: lpa.CertificateProvider.FullName(),
					LpaType:                     appData.Localizer.T(lpa.Type.String()),
					DonorStartPageURL:           appData.PublicURL + page.Paths.Start.Format(),
				}); err != nil {
					return err
				}
			}

			return page.Paths.CertificateProvider.UnableToConfirmIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
		default:
			if err := lpaStoreClient.SendCertificateProviderConfirmIdentity(r.Context(), lpa.LpaUID, certificateProvider); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.OneLoginIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
		}
	}
}
