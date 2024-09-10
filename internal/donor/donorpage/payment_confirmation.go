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
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
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
		if !slices.Contains(provided.PaymentDetails, paymentDetail) {
			provided.PaymentDetails = append(provided.PaymentDetails, paymentDetail)

			if err := eventClient.SendPaymentReceived(r.Context(), event.PaymentReceived{
				UID:       provided.LpaUID,
				PaymentID: payment.PaymentID,
				Amount:    payment.AmountPence.Pence(),
			}); err != nil {
				return err
			}
		}

		if err := notifyClient.SendEmail(r.Context(), payment.Email, notify.PaymentConfirmationEmail{
			DonorFullNamesPossessive: appData.Localizer.Possessive(provided.Donor.FullName()),
			LpaType:                  appData.Localizer.T(provided.Type.String()),
			PaymentCardFullName:      payment.CardDetails.CardholderName,
			LpaReferenceNumber:       provided.LpaUID,
			PaymentReferenceID:       payment.PaymentID,
			PaymentConfirmationDate:  appData.Localizer.FormatDate(payment.SettlementSummary.CaptureSubmitTime),
			AmountPaidWithCurrency:   payment.AmountPence.String(),
		}); err != nil {
			return err
		}

		nextPage := donor.PathTaskList
		if provided.EvidenceDelivery.IsUpload() {
			nextPage = donor.PathEvidenceSuccessfullyUploaded
		} else if provided.EvidenceDelivery.IsPost() {
			nextPage = donor.PathWhatHappensNextPostEvidence
		}

		switch provided.Tasks.PayForLpa {
		case task.PaymentStateInProgress:
			if provided.FeeType.IsFullFee() && provided.FeeAmount() == 0 {
				provided.Tasks.PayForLpa = task.PaymentStateCompleted
			} else {
				provided.Tasks.PayForLpa = task.PaymentStatePending
			}
		case task.PaymentStateApproved, task.PaymentStateDenied:
			if provided.FeeAmount() == 0 {
				provided.Tasks.PayForLpa = task.PaymentStateCompleted
				nextPage = donor.PathTaskList

				if provided.Voucher.Allowed {
					if err := shareCodeSender.SendVoucherAccessCode(r.Context(), provided, appData); err != nil {
						return err
					}

					nextPage = donor.PathWeHaveContactedVoucher
				}

				if provided.Tasks.ConfirmYourIdentityAndSign.IsCompleted() {
					if err := shareCodeSender.SendCertificateProviderPrompt(r.Context(), appData, provided); err != nil {
						return fmt.Errorf("failed to send share code to certificate provider: %w", err)
					}

					if err := eventClient.SendCertificateProviderStarted(r.Context(), event.CertificateProviderStarted{
						UID: provided.LpaUID,
					}); err != nil {
						return fmt.Errorf("failed to send certificate-provider-started event: %w", err)
					}

					if err := lpaStoreClient.SendLpa(r.Context(), provided); err != nil {
						return fmt.Errorf("failed to send to lpastore: %w", err)
					}
				}
			}
		}

		if err := donorStore.Put(r.Context(), provided); err != nil {
			return fmt.Errorf("unable to update lpa in donorStore: %w", err)
		}

		if err := sessionStore.ClearPayment(r, w); err != nil {
			logger.InfoContext(r.Context(), "unable to expire cookie in session", slog.Any("err", err))
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			FeeType:          provided.FeeType,
			PreviousFee:      provided.PreviousFee,
			EvidenceDelivery: provided.EvidenceDelivery,
			NextPage:         nextPage,
		}

		return tmpl(w, data)
	}
}
