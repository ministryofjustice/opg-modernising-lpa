package donorpage

import (
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.DonorIdentityConfirmed() {
			return donor.PathOneLoginIdentityDetails.Redirect(w, r, appData, provided)
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

		provided.DonorIdentityUserData = userData

		if userData.Status.IsFailed() {
			provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateProblem
		} else {
			provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateInProgress
		}

		if userData.Status.IsConfirmed() {
			provided.ProgressSteps.Complete(task.DonorProvedID, time.Now())
		}

		if err := donorStore.Put(r.Context(), provided); err != nil {
			return err
		}

		switch provided.DonorIdentityUserData.Status {
		case identity.StatusFailed:
			return donor.PathRegisterWithCourtOfProtection.Redirect(w, r, appData, provided)
		case identity.StatusInsufficientEvidence:
			return donor.PathUnableToConfirmIdentity.Redirect(w, r, appData, provided)
		default:
			return donor.PathOneLoginIdentityDetails.Redirect(w, r, appData, provided)
		}
	}
}
