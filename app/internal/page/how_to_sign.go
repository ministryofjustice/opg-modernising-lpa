package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
)

type howToSignData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
}

func HowToSign(tmpl template.Template, lpaStore LpaStore, notifyClient NotifyClient, randomCode func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			code := randomCode(4)
			lpa.SignatureCode = code

			emailID, err := notifyClient.Email(r.Context(), notify.Email{
				EmailAddress: lpa.You.Email,
				TemplateID:   notifyClient.TemplateID("MLPA Beta signature code"),
				Personalisation: map[string]string{
					"code": code,
				},
			})
			if err != nil {
				return err
			}
			lpa.SignatureEmailID = emailID

			if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
				return err
			}

			appData.Lang.Redirect(w, r, readYourLpaPath, http.StatusFound)
			return nil
		}

		data := &howToSignData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
