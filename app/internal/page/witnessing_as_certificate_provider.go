package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type witnessingAsCertificateProviderData struct {
	App    AppData
	Errors map[string]string
	Form   *witnessingAsCertificateProviderForm
	Lpa    *Lpa
}

type witnessingAsCertificateProviderForm struct {
	Code string
}

func WitnessingAsCertificateProvider(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &witnessingAsCertificateProviderData{
			App:  appData,
			Lpa:  lpa,
			Form: &witnessingAsCertificateProviderForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if lpa.WitnessCode.HasExpired() {
				data.Errors["witness-code"] = "witnessCodeExpired"
			} else if lpa.WitnessCode.Code != data.Form.Code {
				data.Errors["witness-code"] = "witnessCodeDoesNotMatch"
			}

			if len(data.Errors) == 0 {
				lpa.CPWitnessCodeValidated = true
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, appData.Paths.YouHaveSubmittedYourLpa, http.StatusFound)
			}
		}

		return tmpl(w, data)
	}
}

func readWitnessingAsCertificateProviderForm(r *http.Request) *witnessingAsCertificateProviderForm {
	return &witnessingAsCertificateProviderForm{
		Code: postFormString(r, "witness-code"),
	}
}

func (w *witnessingAsCertificateProviderForm) Validate() map[string]string {
	errors := map[string]string{}

	if w.Code == "" {
		errors["witness-code"] = "enterWitnessCode"
	} else if len(w.Code) < 4 {
		errors["witness-code"] = "witnessCodeTooShort"
	}

	if len(w.Code) > 4 {
		errors["witness-code"] = "witnessCodeTooLong"
	}

	return errors
}
