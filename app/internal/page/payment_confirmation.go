package page

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
)

type paymentConfirmationData struct {
	App              AppData
	Errors           map[string]string
	PaymentReference string
}

func PaymentConfirmation(logger Logger, tmpl template.Template, client pay.PayClient, dataStore DataStore, sessionStore sessions.Store, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve item from data store using key '%s': %s", appData.SessionID, err.Error()))
			return err
		}

		payCookie, err := sessionStore.Get(r, pay.CookieName)

		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve session using key '%s': %s", "pay", err.Error()))
			return err
		}

		paymentId := payCookie.Values[pay.CookiePaymentIdValueKey].(string)
		getPaymentResponse, err := client.GetPayment(paymentId)

		if err != nil {
			logger.Print(fmt.Sprintf("unable to retrieve payment info: %s", err.Error()))
			return err
		}

		lpa.PaymentDetails = PaymentDetails{
			PaymentReference: randomString(12),
			PaymentId:        getPaymentResponse.PaymentId,
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: randomString(12),
		}

		payCookie.Options.MaxAge = -1
		payCookie.Values = map[interface{}]interface{}{pay.CookiePaymentIdValueKey: ""}

		err = sessionStore.Save(r, w, payCookie)

		if err != nil {
			logger.Print(fmt.Sprintf("unable to expire cookie in session: %s", err.Error()))
		}

		lpa.Tasks.PayForLpa = TaskCompleted
		err = dataStore.Put(r.Context(), appData.SessionID, lpa)

		if err != nil {
			logger.Print(fmt.Sprintf("unable to update lpa in dataStore: %s", err.Error()))
			return err
		}

		return tmpl(w, data)
	}
}
