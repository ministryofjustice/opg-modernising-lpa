package voucherpage

import (
	"context"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type vouchFailer func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error

func makeVouchFailer(donorStore DonorStore, notifyClient NotifyClient, appPublicURL string) vouchFailer {
	return func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error {
		donor, err := donorStore.GetAny(ctx)
		if err != nil {
			return fmt.Errorf("could not get donor: %w", err)
		}

		email := notify.VouchingFailedAttemptEmail{
			Greeting:          notifyClient.EmailGreeting(lpa),
			VoucherFullName:   provided.FullName(),
			DonorStartPageURL: appPublicURL + page.PathStart.Format(),
		}

		if err := notifyClient.SendActorEmail(ctx, lpa.Donor.ContactLanguagePreference, lpa.CorrespondentEmail(), lpa.LpaUID, email); err != nil {
			return fmt.Errorf("could not send email: %w", err)
		}

		if err := donorStore.FailVoucher(ctx, donor, provided.SK); err != nil {
			return fmt.Errorf("could not fail vouch: %w", err)
		}

		return nil
	}
}
