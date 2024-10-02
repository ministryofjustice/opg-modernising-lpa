package donorpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, donorStore DonorStore, scheduledStore ScheduledStore) Handler {
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

		userData, err := oneLoginClient.ParseIdentityClaim(userInfo)
		if err != nil {
			return err
		}

		provided.IdentityUserData = userData

		if userData.Status.IsFailed() {
			provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateProblem
		} else {
			provided.Tasks.ConfirmYourIdentityAndSign = task.IdentityStateInProgress
		}

		if err := donorStore.Put(r.Context(), provided); err != nil {
			return err
		}

		switch provided.IdentityUserData.Status {
		case identity.StatusFailed:
			return donor.PathRegisterWithCourtOfProtection.Redirect(w, r, appData, provided)
		case identity.StatusInsufficientEvidence:
			return donor.PathUnableToConfirmIdentity.Redirect(w, r, appData, provided)
		default:
			if err := scheduledStore.Put(r.Context(), scheduled.Event{
				At:                userData.RetrievedAt.AddDate(0, 6, 0),
				Action:            scheduled.ActionExpireDonorIdentity,
				TargetLpaKey:      provided.PK,
				TargetLpaOwnerKey: provided.SK,
			}); err != nil {
				return err
			}

			return donor.PathOneLoginIdentityDetails.Redirect(w, r, appData, provided)
		}
	}
}
