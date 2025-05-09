package certificateproviderpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreClient LpaStoreClient, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			return certificateprovider.PathIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
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

		userData, err := oneLoginClient.ParseIdentityClaim(userInfo)
		if err != nil {
			return err
		}

		certificateProvider.IdentityUserData = userData

		if userData.Status.IsConfirmed() && !certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			certificateProvider.IdentityDetailsMismatched = true
			certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStatePending
		} else {
			certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
		}

		if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
			return err
		}

		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			if err := lpaStoreClient.SendCertificateProviderConfirmIdentity(r.Context(), lpa.LpaUID, certificateProvider); err != nil {
				return err
			}

			return certificateprovider.PathIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		if certificateProvider.IdentityUserData.Status.IsConfirmed() || certificateProvider.IdentityUserData.Status.IsFailed() {
			if err := eventClient.SendIdentityCheckMismatched(r.Context(), event.IdentityCheckMismatched{
				LpaUID:   lpa.LpaUID,
				ActorUID: certificateProvider.UID,
				Provided: event.IdentityCheckMismatchedDetails{
					FirstNames:  lpa.CertificateProvider.FirstNames,
					LastName:    lpa.CertificateProvider.LastName,
					DateOfBirth: certificateProvider.DateOfBirth,
				},
				Verified: event.IdentityCheckMismatchedDetails{
					FirstNames:  userData.FirstNames,
					LastName:    userData.LastName,
					DateOfBirth: userData.DateOfBirth,
				},
			}); err != nil {
				return err
			}
		}

		return certificateprovider.PathIdentityDetails.Redirect(w, r, appData, certificateProvider.LpaID)
	}
}
