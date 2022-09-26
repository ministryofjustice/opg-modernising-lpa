package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"

	"github.com/gorilla/sessions"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
)

type paymentConfirmationData struct {
	App              AppData
	Errors           map[string]string
	PaymentReference string
}

func PaymentConfirmation(logger Logger, tmpl template.Template, client pay.PayClient, dataStore DataStore, sessionStore sessions.Store, random random.RandomGenerator) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		payCookie, _ := sessionStore.Get(r, pay.CookieName)

		paymentId := payCookie.Values[pay.CookiePaymentIdValueKey].(string)
		getPaymentResponse, _ := client.GetPayment(paymentId)

		paymentReference := random.String(12)
		lpa.PaymentDetails = PaymentDetails{
			PaymentReference: paymentReference,
			PaymentId:        getPaymentResponse.PaymentId,
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: paymentReference,
		}

		payCookie.Options.MaxAge = -1
		payCookie.Values = map[interface{}]interface{}{pay.CookiePaymentIdValueKey: ""}

		sessionStore.Save(r, w, payCookie)

		lpa.Tasks.PayForLpa = TaskCompleted
		dataStore.Put(r.Context(), appData.SessionID, lpa)

		return tmpl(w, data)
	}
}
