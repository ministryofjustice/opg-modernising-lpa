package donor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type aboutPaymentData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
}

func AboutPayment(logger Logger, tmpl template.Template, sessionStore sessions.Store, payClient PayClient, appPublicUrl string, randomString func(int) string, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &aboutPaymentData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
		}

		if r.Method == http.MethodPost {
			createPaymentBody := pay.CreatePaymentBody{
				Amount:      page.CostOfLpaPence,
				Reference:   randomString(12),
				Description: "Property and Finance LPA",
				ReturnUrl:   appPublicUrl + appData.BuildUrl(page.Paths.PaymentConfirmation),
				Email:       "a@b.com",
				Language:    appData.Lang.String(),
			}

			resp, err := payClient.CreatePayment(createPaymentBody)
			if err != nil {
				logger.Print(fmt.Sprintf("Error creating payment: %s", err.Error()))
				return err
			}

			if err = sesh.SetPayment(sessionStore, r, w, &sesh.PaymentSession{PaymentID: resp.PaymentId}); err != nil {
				return err
			}

			nextUrl := resp.Links["next_url"].Href
			// If URL matches expected domain for GOV UK PAY redirect there. If not,
			// redirect to the confirmation code and carry on with flow.
			if strings.HasPrefix(nextUrl, pay.PaymentPublicServiceUrl) {
				http.Redirect(w, r, nextUrl, http.StatusFound)
			} else {
				appData.Redirect(w, r, lpa, page.Paths.PaymentConfirmation)
			}

		}

		return tmpl(w, data)
	}
}
