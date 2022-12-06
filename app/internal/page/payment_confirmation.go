package page

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
)

type paymentConfirmationData struct {
	App              AppData
	Errors           map[string]string
	PaymentReference string
	Continue         string
}

func PaymentConfirmation(logger Logger, tmpl template.Template, client PayClient, lpaStore LpaStore, sessionStore sessions.Store) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		payCookie, err := sessionStore.Get(r, PayCookieName)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve session using key '%s': %s", "pay", err.Error()))
			return err
		}

		paymentId := payCookie.Values[PayCookiePaymentIdValueKey].(string)

		payment, err := client.GetPayment(paymentId)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve payment info: %s", err.Error()))
			return err
		}

		lpa.PaymentDetails = PaymentDetails{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentId,
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			Continue:         appData.Paths.TaskList,
		}

		payCookie.Options.MaxAge = -1
		payCookie.Values = map[interface{}]interface{}{PayCookiePaymentIdValueKey: ""}

		if err := sessionStore.Save(r, w, payCookie); err != nil {
			logger.Print(fmt.Sprintf("unable to expire cookie in session: %s", err.Error()))
		}

		lpa.Tasks.PayForLpa = TaskCompleted

		if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in dataStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
