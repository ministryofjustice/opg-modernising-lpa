package voucherpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, voucherStore VoucherStore, lpaStoreResolvingService LpaStoreResolvingService, notifyClient NotifyClient, appPublicURL string, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
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

		provided.IdentityUserData = userData
		provided.IdentityUserData.VouchedFor = true

		if provided.NameMatches(lpa).IsNone() {
			provided.Tasks.ConfirmYourIdentity = task.StateCompleted
		} else {
			provided.Tasks.ConfirmYourIdentity = task.StateInProgress
		}

		if err := voucherStore.Put(r.Context(), provided); err != nil {
			return err
		}

		if !provided.IdentityConfirmed() {
			if !lpa.SignedAt.IsZero() {
				if err = notifyClient.SendActorEmail(r.Context(), lpa.CorrespondentEmail(), lpa.LpaUID, notify.VoucherFailedIdentityCheckEmail{
					Greeting:          notifyClient.EmailGreeting(lpa),
					DonorFullName:     lpa.Donor.FullName(),
					VoucherFullName:   lpa.Voucher.FullName(),
					LpaType:           appData.Localizer.T(lpa.Type.String()),
					DonorStartPageURL: appPublicURL + page.PathStart.Format(),
				}); err != nil {
					return err
				}
			}

			donor, err := donorStore.GetAny(r.Context())
			if err != nil {
				return err
			}

			donor.FailedVouchAttempts++

			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			return voucher.PathUnableToConfirmIdentity.Redirect(w, r, appData, appData.LpaID)
		}

		if provided.Tasks.ConfirmYourIdentity.IsCompleted() {
			return voucher.PathOneLoginIdentityDetails.Redirect(w, r, appData, appData.LpaID)
		}

		return voucher.PathConfirmAllowedToVouch.Redirect(w, r, appData, appData.LpaID)
	}
}
