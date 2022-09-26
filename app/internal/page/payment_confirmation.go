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

		data := &paymentConfirmationData{
			App: appData,
		}

		payCookie, _ := sessionStore.Get(r, "pay")

		getPaymentResponse, _ := client.GetPayment(payCookie.Values["paymentId"].(string))

		lpa.PaymentDetails = PaymentDetails{
			PaymentReference: random.String(12),
			PaymentId:        getPaymentResponse.PaymentId,
		}

		dataStore.Put(r.Context(), appData.SessionID, lpa)

		payCookie.Options.MaxAge = -1
		payCookie.Values = map[interface{}]interface{}{"paymentId": ""}

		sessionStore.Save(r, w, payCookie)

		return tmpl(w, data)
	}
}
