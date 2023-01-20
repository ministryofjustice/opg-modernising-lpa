package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type witnessingYourSignatureData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
}

func WitnessingYourSignature(tmpl template.Template, lpaStore LpaStore, notifyClient NotifyClient, randomCode func(int) string, now func() time.Time) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			code := randomCode(4)
			lpa.WitnessCode = WitnessCode{Code: code, Created: now()}

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

			return appData.Redirect(w, r, lpa, Paths.WitnessingAsCertificateProvider)
		}

		data := &witnessingYourSignatureData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
