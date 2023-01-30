package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsCertificateProviderData struct {
	App    AppData
	Errors validation.List
	Form   *witnessingAsCertificateProviderForm
	Lpa    *Lpa
}

func WitnessingAsCertificateProvider(tmpl template.Template, lpaStore LpaStore, now func() time.Time) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
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
				data.Errors.Add("witness-code", "witnessCodeExpired")
			} else if lpa.WitnessCode.Code != data.Form.Code {
				data.Errors.Add("witness-code", "witnessCodeDoesNotMatch")
			}

			if data.Errors.None() {
				lpa.CPWitnessCodeValidated = true
				lpa.Submitted = now()
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.YouHaveSubmittedYourLpa)
			}
		}

		return tmpl(w, data)
	}
}

type witnessingAsCertificateProviderForm struct {
	Code string
}

func readWitnessingAsCertificateProviderForm(r *http.Request) *witnessingAsCertificateProviderForm {
	return &witnessingAsCertificateProviderForm{
		Code: postFormString(r, "witness-code"),
	}
}

func (w *witnessingAsCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	if w.Code == "" {
		errors.Add("witness-code", "enterWitnessCode")
	} else if len(w.Code) < 4 {
		errors.Add("witness-code", "witnessCodeTooShort")
	}

	if len(w.Code) > 4 {
		errors.Add("witness-code", "witnessCodeTooLong")
	}

	return errors
}
