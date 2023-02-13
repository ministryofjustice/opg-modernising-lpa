package donor

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type aboutPaymentData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
}

func AboutPayment(logger page.Logger, tmpl template.Template, sessionStore sessions.Store, payClient page.PayClient, appPublicUrl string, randomString func(int) string, lpaStore page.LpaStore) page.Handler {
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

			nextUrl := resp.Links["next_url"].Href

			secureCookies := strings.HasPrefix(nextUrl, "https:")

			// TODO move to sesh
			cookieOptions := &sessions.Options{
				Path: "/",
				// A payment can be resumed up to 90 minutes after creation
				MaxAge:   int(time.Minute * 90 / time.Second),
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   secureCookies,
			}

			session := sessions.NewSession(sessionStore, page.PayCookieName)
			session.Values = map[interface{}]interface{}{
				page.PayCookiePaymentIdValueKey: resp.PaymentId,
			}
			session.Options = cookieOptions

			if err = sessionStore.Save(r, w, session); err != nil {
				return err
			}

			// If URL matches expected domain for GOV UK PAY redirect there. If not, redirect to the confirmation code and carry on with flow.
			if strings.HasPrefix(nextUrl, pay.PaymentPublicServiceUrl) {
				http.Redirect(w, r, nextUrl, http.StatusFound)
			} else {
				appData.Redirect(w, r, lpa, page.Paths.PaymentConfirmation)
			}

		}

		return tmpl(w, data)
	}
}
