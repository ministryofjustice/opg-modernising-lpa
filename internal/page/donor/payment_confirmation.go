package donor

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type paymentConfirmationData struct {
	App              page.AppData
	Errors           validation.List
	PaymentReference string
	FeeType          pay.FeeType
	PreviousFee      pay.PreviousFee
	EvidenceDelivery pay.EvidenceDelivery
	NextPage         page.LpaPath
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore SessionStore, shareCodeSender ShareCodeSender, lpaStoreClient LpaStoreClient, eventClient EventClient, notifyClient NotifyClient) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		paymentSession, err := sessionStore.Payment(r)
		if err != nil {
			return err
		}

		payment, err := payClient.GetPayment(r.Context(), paymentSession.PaymentID)
		if err != nil {
			return fmt.Errorf("unable to retrieve payment info: %w", err)
		}

		if payment.State.Status != "success" {
			return errors.New("TODO: we need to give some options")
		}

		paymentDetail := actor.Payment{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentID,
			Amount:           payment.AmountPence.Int(),
		}
		if !slices.Contains(donor.PaymentDetails, paymentDetail) {
			donor.PaymentDetails = append(donor.PaymentDetails, paymentDetail)

			if err := eventClient.SendPaymentReceived(r.Context(), event.PaymentReceived{
				UID:       donor.LpaUID,
				PaymentID: payment.PaymentID,
				Amount:    payment.AmountPence.Int(),
			}); err != nil {
				return err
			}
		}

		if err := notifyClient.SendEmail(r.Context(), payment.Email, notify.PaymentConfirmationEmail{
			DonorFullNamesPossessive: appData.Localizer.Possessive(donor.Donor.FullName()),
			LpaType:                  appData.Localizer.T(donor.Type.String()),
			PaymentCardFullName:      payment.CardDetails.CardholderName,
			LpaReferenceNumber:       donor.LpaUID,
			PaymentReferenceID:       payment.PaymentID,
			PaymentConfirmationDate:  appData.Localizer.FormatDate(payment.SettlementSummary.CaptureSubmitTime),
			AmountPaidWithCurrency:   payment.AmountPence.Pounds(),
		}); err != nil {
			return err
		}

		nextPage := page.Paths.TaskList
		if donor.EvidenceDelivery.IsUpload() {
			nextPage = page.Paths.EvidenceSuccessfullyUploaded
		} else if donor.EvidenceDelivery.IsPost() {
			nextPage = page.Paths.WhatHappensNextPostEvidence
		}

		switch donor.Tasks.PayForLpa {
		case actor.PaymentTaskInProgress:
			if donor.FeeType.IsFullFee() && donor.FeeAmount() == 0 {
				donor.Tasks.PayForLpa = actor.PaymentTaskCompleted
			} else {
				donor.Tasks.PayForLpa = actor.PaymentTaskPending
			}
		case actor.PaymentTaskApproved, actor.PaymentTaskDenied:
			if donor.FeeAmount() == 0 {
				donor.Tasks.PayForLpa = actor.PaymentTaskCompleted
				nextPage = page.Paths.TaskList

				if donor.Voucher.Allowed {
					// TODO: MLPAB-1897 send code to donor and MLPAB-1899 contact voucher
					nextPage = page.Paths.WeHaveContactedVoucher
				}

				if donor.Tasks.ConfirmYourIdentityAndSign.IsCompleted() {
					if err := shareCodeSender.SendCertificateProviderPrompt(r.Context(), appData, donor); err != nil {
						return fmt.Errorf("failed to send share code to certificate provider: %w", err)
					}

					if err := lpaStoreClient.SendLpa(r.Context(), donor); err != nil {
						return fmt.Errorf("failed to send to lpastore: %w", err)
					}
				}
			}
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			return fmt.Errorf("unable to update lpa in donorStore: %w", err)
		}

		if err := sessionStore.ClearPayment(r, w); err != nil {
			logger.InfoContext(r.Context(), "unable to expire cookie in session", slog.Any("err", err))
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			FeeType:          donor.FeeType,
			PreviousFee:      donor.PreviousFee,
			EvidenceDelivery: donor.EvidenceDelivery,
			NextPage:         nextPage,
		}

		return tmpl(w, data)
	}
}
