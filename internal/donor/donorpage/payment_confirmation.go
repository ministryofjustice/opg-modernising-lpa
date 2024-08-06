package donorpage

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type paymentConfirmationData struct {
	App              appcontext.Data
	Errors           validation.List
	PaymentReference string
	FeeType          pay.FeeType
	PreviousFee      pay.PreviousFee
	EvidenceDelivery pay.EvidenceDelivery
	NextPage         donor.Path
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore SessionStore, shareCodeSender ShareCodeSender, lpaStoreClient LpaStoreClient, eventClient EventClient, notifyClient NotifyClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
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

		paymentDetail := donordata.Payment{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentID,
			Amount:           payment.AmountPence.Pence(),
		}
		if !slices.Contains(donor.PaymentDetails, paymentDetail) {
			donor.PaymentDetails = append(donor.PaymentDetails, paymentDetail)

			if err := eventClient.SendPaymentReceived(r.Context(), event.PaymentReceived{
				UID:       donor.LpaUID,
				PaymentID: payment.PaymentID,
				Amount:    payment.AmountPence.Pence(),
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
			AmountPaidWithCurrency:   payment.AmountPence.String(),
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
		case task.PaymentStateInProgress:
			if donor.FeeType.IsFullFee() && donor.FeeAmount() == 0 {
				donor.Tasks.PayForLpa = task.PaymentStateCompleted
			} else {
				donor.Tasks.PayForLpa = task.PaymentStatePending
			}
		case task.PaymentStateApproved, task.PaymentStateDenied:
			if donor.FeeAmount() == 0 {
				donor.Tasks.PayForLpa = task.PaymentStateCompleted
				nextPage = page.Paths.TaskList

				if donor.Voucher.Allowed {
					// TODO: MLPAB-1897 send code to donor and MLPAB-1899 contact voucher
					nextPage = page.Paths.WeHaveContactedVoucher
				}

				if donor.Tasks.ConfirmYourIdentityAndSign.IsCompleted() {
					if err := shareCodeSender.SendCertificateProviderPrompt(r.Context(), appData, donor); err != nil {
						return fmt.Errorf("failed to send share code to certificate provider: %w", err)
					}

					if err := eventClient.SendCertificateProviderStarted(r.Context(), event.CertificateProviderStarted{
						UID: donor.LpaUID,
					}); err != nil {
						return fmt.Errorf("failed to send certificate-provider-started event: %w", err)
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
