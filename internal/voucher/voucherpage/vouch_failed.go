package voucherpage

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type failVouch func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error

func vouchFailed(donorStore DonorStore, notifyClient NotifyClient, appPublicURL string) failVouch {
	return func(ctx context.Context, provided *voucherdata.Provided, lpa *lpadata.Lpa) error {
		donor, err := donorStore.GetAny(ctx)
		if err != nil {
			return err
		}

		donor.FailedVouchAttempts++
		donor.WantVoucher = form.No
		donor.Voucher = donordata.Voucher{}

		email := notify.VouchingFailedAttemptEmail{
			Greeting:          notifyClient.EmailGreeting(lpa),
			VoucherFullName:   provided.FullName(),
			DonorStartPageURL: appPublicURL + page.PathStart.Format(),
		}

		if err := notifyClient.SendActorEmail(ctx, lpa.Donor.Email, lpa.LpaUID, email); err != nil {
			return err
		}

		if err := donorStore.Put(ctx, donor); err != nil {
			return err
		}

		return nil
	}
}
