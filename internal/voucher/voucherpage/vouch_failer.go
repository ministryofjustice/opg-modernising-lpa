package voucherpage

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type vouchFailer func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error

func makeVouchFailer(donorStore DonorStore, notifyClient NotifyClient, donorStartURL string) vouchFailer {
	return func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error {
		donor, err := donorStore.GetAny(ctx)
		if err != nil {
			return fmt.Errorf("could not get donor: %w", err)
		}

		email := notify.VouchingFailedAttemptEmail{
			Greeting:          notifyClient.EmailGreeting(lpa),
			VoucherFullName:   provided.FullName(),
			DonorStartPageURL: donorStartURL,
		}

		if err := notifyClient.SendActorEmail(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, email); err != nil {
			return fmt.Errorf("could not send email: %w", err)
		}

		if err := donorStore.FailVoucher(ctx, donor); err != nil {
			return fmt.Errorf("could not fail vouch: %w", err)
		}

		return nil
	}
}
