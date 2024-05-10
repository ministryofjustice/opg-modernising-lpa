package donor

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore SessionStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		paymentSession, err := sessionStore.Payment(r)
		if err != nil {
			return err
		}

		paymentId := paymentSession.PaymentID

		payment, err := payClient.GetPayment(r.Context(), paymentId)
		if err != nil {
			return fmt.Errorf("unable to retrieve payment info: %w", err)
		}

		donor.PaymentDetails = append(donor.PaymentDetails, actor.Payment{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentId,
			Amount:           payment.Amount,
		})

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			FeeType:          donor.FeeType,
			PreviousFee:      donor.PreviousFee,
			EvidenceDelivery: donor.EvidenceDelivery,
		}

		if err := sessionStore.ClearPayment(r, w); err != nil {
			logger.InfoContext(r.Context(), "unable to expire cookie in session", slog.Any("err", err))
		}

		if donor.FeeType.IsFullFee() {
			donor.Tasks.PayForLpa = actor.PaymentTaskCompleted
		} else {
			donor.Tasks.PayForLpa = actor.PaymentTaskPending
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			return fmt.Errorf("unable to update lpa in donorStore: %w", err)
		}

		return tmpl(w, data)
	}
}
