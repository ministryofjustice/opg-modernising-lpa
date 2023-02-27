package donor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type paymentConfirmationData struct {
	App              page.AppData
	Errors           validation.List
	PaymentReference string
}

func PaymentConfirmation(logger Logger, tmpl template.Template, payClient PayClient, lpaStore LpaStore, sessionStore sessions.Store, shareCodeSender ShareCodeSender) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

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

		if err := shareCodeSender.Send(r.Context(), notify.CertificateProviderInviteEmail, appData, true, lpa); err != nil {
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

		lpa.Tasks.PayForLpa = page.TaskCompleted

		if err := lpaStore.Put(r.Context(), lpa); err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in dataStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
