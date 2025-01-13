package voucherpage

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type yourDeclarationData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *yourDeclarationForm
	Lpa     *lpadata.Lpa
	Voucher *voucherdata.Provided
}

func YourDeclaration(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore, donorStore DonorStore, notifyClient NotifyClient, lpaStoreClient LpaStoreClient, now func() time.Time, appPublicURL string) Handler {
	sendNotification := func(ctx context.Context, lpa *lpadata.Lpa, provided *voucherdata.Provided) error {
		if lpa.Donor.Mobile != "" {
			if !lpa.SignedForDonor() {
				return notifyClient.SendActorSMS(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, notify.VoucherHasConfirmedDonorIdentitySMS{
					VoucherFullName:   provided.FullName(),
					DonorFullName:     lpa.Donor.FullName(),
					DonorStartPageURL: appPublicURL + page.PathStart.Format(),
				})
			}

			return notifyClient.SendActorSMS(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, notify.VoucherHasConfirmedDonorIdentityOnSignedLpaSMS{
				VoucherFullName:   provided.FullName(),
				DonorStartPageURL: appPublicURL + page.PathStart.Format(),
			})
		}

		if !lpa.SignedForDonor() {
			return notifyClient.SendActorEmail(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, notify.VoucherHasConfirmedDonorIdentityEmail{
				VoucherFullName:   provided.FullName(),
				DonorFullName:     lpa.Donor.FullName(),
				DonorStartPageURL: appPublicURL + page.PathStart.Format(),
			})
		}

		return notifyClient.SendActorEmail(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, notify.VoucherHasConfirmedDonorIdentityOnSignedLpaEmail{
			VoucherFullName:   provided.FullName(),
			DonorFullName:     lpa.Donor.FullName(),
			DonorStartPageURL: appPublicURL + page.PathStart.Format(),
		})
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		if !provided.SignedAt.IsZero() {
			return voucher.PathThankYou.Redirect(w, r, appData, appData.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("error resolving lpa: %w", err)
		}

		data := &yourDeclarationData{
			App:     appData,
			Form:    &yourDeclarationForm{},
			Lpa:     lpa,
			Voucher: provided,
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDeclarationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if err := sendNotification(r.Context(), lpa, provided); err != nil {
					return fmt.Errorf("error sending notification: %w", err)
				}

				donor, err := donorStore.GetAny(r.Context())
				if err != nil {
					return fmt.Errorf("error getting donor: %w", err)
				}

				donor.IdentityUserData = identity.UserData{
					Status:         identity.StatusConfirmed,
					FirstNames:     donor.Donor.FirstNames,
					LastName:       donor.Donor.LastName,
					DateOfBirth:    donor.Donor.DateOfBirth,
					CurrentAddress: donor.Donor.Address,
					CheckedAt:      now(),
				}
				donor.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return fmt.Errorf("error updating donor: %w", err)
				}

				provided.SignedAt = now()
				provided.Tasks.SignTheDeclaration = task.StateCompleted
				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return fmt.Errorf("error updating voucher: %w", err)
				}

				if lpa.Submitted {
					if err := lpaStoreClient.SendDonorConfirmIdentity(r.Context(), donor); err != nil {
						return fmt.Errorf("error sending donor identity confirmation: %w", err)
					}
				}

				return voucher.PathThankYou.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type yourDeclarationForm struct {
	Confirm bool
}

func readYourDeclarationForm(r *http.Request) *yourDeclarationForm {
	return &yourDeclarationForm{
		Confirm: page.PostFormString(r, "confirm") == "1",
	}
}

func (f *yourDeclarationForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("confirm", "youMustSelectTheBoxToVouch", f.Confirm,
		validation.Selected().CustomError())

	return errors
}
