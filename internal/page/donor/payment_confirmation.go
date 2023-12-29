package donor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
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

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore sessions.Store) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		paymentSession, err := sesh.Payment(sessionStore, r)
		if err != nil {
			return err
		}

		paymentId := paymentSession.PaymentID

		payment, err := payClient.GetPayment(r.Context(), paymentId)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve payment info: %s", err.Error()))
			return err
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

		if err := sesh.ClearPayment(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire cookie in session: %s", err.Error()))
		}

		if donor.FeeType.IsFullFee() {
			donor.Tasks.PayForLpa = actor.PaymentTaskCompleted
		} else {
			donor.Tasks.PayForLpa = actor.PaymentTaskPending
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in donorStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
