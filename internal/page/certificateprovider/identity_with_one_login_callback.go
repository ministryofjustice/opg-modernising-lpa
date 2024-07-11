package certificateprovider

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) page.Handler {
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
			return page.Paths.CertificateProvider.OneloginIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
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

		if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
			return err
		}

		switch certificateProvider.IdentityUserData.Status {
		case identity.StatusFailed, identity.StatusInsufficientEvidence, identity.StatusUnknown:
			return page.Paths.CertificateProvider.UnableToConfirmIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
		default:
			return page.Paths.CertificateProvider.OneloginIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
		}
	}
}
