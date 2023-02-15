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
	Continue         string
}

func PaymentConfirmation(logger page.Logger, tmpl template.Template, payClient page.PayClient, notifyClient page.NotifyClient, lpaStore page.LpaStore, sessionStore sessions.Store, appPublicURL string, dataStore page.DataStore, randomString func(int) string) page.Handler {
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

		shareCode := randomString(12)

		if err := dataStore.Put(r.Context(), "SHARECODE#"+shareCode, "#METADATA#"+shareCode, page.ShareCodeData{
			SessionID: appData.SessionID,
			LpaID:     appData.LpaID,
		}); err != nil {
			return err
		}

		if _, err := notifyClient.Email(r.Context(), notify.Email{
			TemplateID:   notifyClient.TemplateID(notify.CertificateProviderInviteEmail),
			EmailAddress: lpa.CertificateProvider.Email,
			Personalisation: map[string]string{
				"link": fmt.Sprintf("%s%s?share-code=%s", appPublicURL, page.Paths.CertificateProviderStart, shareCode),
			},
		}); err != nil {
			return fmt.Errorf("error email certificate provider after payment: %w", err)
		}

		lpa.PaymentDetails = page.PaymentDetails{
			PaymentReference: payment.Reference,
			PaymentId:        payment.PaymentId,
		}

		data := &paymentConfirmationData{
			App:              appData,
			PaymentReference: payment.Reference,
			Continue:         appData.Paths.TaskList,
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
