package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if donor.DonorIdentityConfirmed() {
			return page.Paths.OneloginIdentityDetails.Redirect(w, r, appData, donor)
		}

		if r.FormValue("error") == "access_denied" {
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

		donor.DonorIdentityUserData = userData

		if userData.Status.IsFailed() {
			donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskProblem
		} else {
			donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskInProgress
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			return err
		}

		switch donor.DonorIdentityUserData.Status {
		case identity.StatusFailed:
			return page.Paths.RegisterWithCourtOfProtection.Redirect(w, r, appData, donor)
		case identity.StatusInsufficientEvidence:
			return page.Paths.UnableToConfirmIdentity.Redirect(w, r, appData, donor)
		default:
			return page.Paths.OneloginIdentityDetails.Redirect(w, r, appData, donor)
		}
	}
}
