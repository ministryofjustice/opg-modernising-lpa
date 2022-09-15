package page

import (
	"net/http"
	"strings"

	"github.com/gorilla/sessions"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"

	"github.com/ministryofjustice/opg-go-common/template"
)

type aboutPaymentData struct {
	App    AppData
	Errors map[string]string
}

func AboutPayment(tmpl template.Template, store sessions.Store) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &aboutPaymentData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			payClient, err := pay.New("http://pay-mock:4010", "fake-key", http.DefaultClient)

			if err != nil {
				return err
			}

			createPaymentBody := pay.CreatePaymentBody{
				Amount:      0,
				Reference:   "abc",
				Description: "A payment",
				ReturnUrl:   "/url/1",
				Email:       "a@b.com",
				Language:    "en",
			}

			resp, err := payClient.CreatePayment(createPaymentBody)

			if err != nil {
				return err
			}

			//secureCookies := strings.HasPrefix(appPublicURL, "https:")

			cookieOptions := &sessions.Options{
				Path:     "/",
				MaxAge:   24 * 60 * 60,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   false,
			}

			session := sessions.NewSession(store, "pay")
			session.Values = map[interface{}]interface{}{
				"paymentId": resp.PaymentId,
			}
			session.Options = cookieOptions

			session.Values = map[interface{}]interface{}{"paymentId": resp.PaymentId}
			if err := store.Save(r, w, session); err != nil {
				return err
			}

			// If URL matches expected domain for GOV UK PAY redirect there. If not, redirect to the confirmation code and carry on with flow.
			redirectUrl := resp.Links["next_url"].Href
			if strings.Contains(redirectUrl, "https://publicapi.payments.service.gov.uk/") {
				http.Redirect(w, r, redirectUrl, http.StatusFound)
			} else {
				http.Redirect(w, r, "/payment-confirmation", http.StatusFound)
			}

		}

		return tmpl(w, data)
	}
}
