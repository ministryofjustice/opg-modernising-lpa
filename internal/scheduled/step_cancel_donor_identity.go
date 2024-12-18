package scheduled

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func (r *Runner) stepCancelDonorIdentity(ctx context.Context, row *Event) error {
	provided, err := r.donorStore.One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey)
	if err != nil {
		return fmt.Errorf("error retrieving donor: %w", err)
	}

	if !provided.IdentityUserData.Status.IsConfirmed() || !provided.SignedAt.IsZero() {
		return errStepIgnored
	}

	provided.IdentityUserData = identity.UserData{Status: identity.StatusExpired}
	provided.Tasks.ConfirmYourIdentity = task.IdentityStateNotStarted

	if err := r.notifyClient.SendActorEmail(ctx, notify.ToDonor(provided), provided.LpaUID, notify.DonorIdentityCheckExpiredEmail{}); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	if err := r.donorStore.Put(ctx, provided); err != nil {
		return fmt.Errorf("error updating donor: %w", err)
	}

	return nil
}
