package donorpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, donorStore DonorStore, scheduledStore ScheduledStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.DonorIdentityConfirmed() {
			return donor.PathIdentityDetails.Redirect(w, r, appData, provided)
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
			provided.Tasks.ConfirmYourIdentity = task.IdentityStateProblem
		} else if provided.DonorIdentityConfirmed() {
			provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
		}

		if (!provided.WitnessedByCertificateProviderAt.IsZero() && !provided.DonorIdentityConfirmed()) || provided.IdentityUserData.Status.IsFailed() {
			if err := eventClient.SendIdentityCheckMismatched(r.Context(), event.IdentityCheckMismatched{
				LpaUID:   provided.LpaUID,
				ActorUID: provided.Donor.UID,
				Provided: event.IdentityCheckMismatchedDetails{
					FirstNames:  provided.Donor.FirstNames,
					LastName:    provided.Donor.LastName,
					DateOfBirth: provided.Donor.DateOfBirth,
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

		if err := donorStore.Put(r.Context(), provided); err != nil {
			return err
		}

		switch provided.IdentityUserData.Status {
		case identity.StatusFailed:
			return donor.PathRegisterWithCourtOfProtection.Redirect(w, r, appData, provided)
		case identity.StatusInsufficientEvidence:
			return donor.PathUnableToConfirmIdentity.Redirect(w, r, appData, provided)
		default:
			if err := scheduledStore.Create(r.Context(), scheduled.Event{
				At:                userData.CheckedAt.AddDate(0, 6, 0),
				Action:            scheduled.ActionExpireDonorIdentity,
				TargetLpaKey:      provided.PK,
				TargetLpaOwnerKey: provided.SK,
				LpaUID:            provided.LpaUID,
			}); err != nil {
				return err
			}

			return donor.PathIdentityDetails.Redirect(w, r, appData, provided)
		}
	}
}
