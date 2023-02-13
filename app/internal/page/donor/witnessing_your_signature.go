package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingYourSignatureData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WitnessingYourSignature(tmpl template.Template, lpaStore page.LpaStore, notifyClient page.NotifyClient, randomCode func(int) string, now func() time.Time) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			code := randomCode(4)
			lpa.WitnessCode = page.WitnessCode{Code: code, Created: now()}

			smsID, err := notifyClient.Sms(r.Context(), notify.Sms{
				PhoneNumber: lpa.CertificateProvider.Mobile,
				TemplateID:  notifyClient.TemplateID(notify.SignatureCodeSms),
				Personalisation: map[string]string{
					"code": code,
				},
			})

			if err != nil {
				return err
			}

			lpa.SignatureSmsID = smsID

			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.WitnessingAsCertificateProvider)
		}

		data := &witnessingYourSignatureData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
