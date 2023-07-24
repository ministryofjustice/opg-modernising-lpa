package donor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type paymentConfirmationData struct {
	App              page.AppData
	Errors           validation.List
	PaymentReference string
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, donorStore DonorStore, sessionStore sessions.Store) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		paymentSession, err := sesh.Payment(sessionStore, r)
		if err != nil {
			return err
		}

		paymentId := paymentSession.PaymentID

		payment, err := payClient.GetPayment(paymentId)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve payment info: %s", err.Error()))
			return err
		}

		lpa.PaymentDetails = page.PaymentDetails{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentId,
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
		}

		if err := sesh.ClearPayment(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire cookie in session: %s", err.Error()))
		}

		if lpa.FeeType.IsFullFee() {
			lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
		} else {
			lpa.Tasks.PayForLpa = actor.PaymentTaskPending
		}

		if err := donorStore.Put(r.Context(), lpa); err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in donorStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
