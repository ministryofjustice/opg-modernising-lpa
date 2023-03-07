package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsCertificateProviderData struct {
	App    page.AppData
	Errors validation.List
	Form   *witnessingAsCertificateProviderForm
	Lpa    *page.Lpa
}

func WitnessingAsCertificateProvider(tmpl template.Template, lpaStore LpaStore, shareCodeSender ShareCodeSender, now func() time.Time) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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

			if lpa.WitnessCodeLimiter == nil {
				lpa.WitnessCodeLimiter = page.NewLimiter(time.Minute, 5, 10)
			}

			if !lpa.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := lpa.WitnessCodes.Find(data.Form.Code)
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired() {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				lpa.WitnessCodeLimiter = nil
				lpa.CPWitnessCodeValidated = true
				lpa.Submitted = now()
			}

			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if data.Errors.None() {
				if lpa.CertificateProviderOneLoginUserData.OK {
					if err := shareCodeSender.Send(r.Context(), notify.CertificateProviderReturnEmail, appData, false, lpa); err != nil {
						return err
					}
				}

				return appData.Redirect(w, r, lpa, page.Paths.YouHaveSubmittedYourLpa)
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
		Code: page.PostFormString(r, "witness-code"),
	}
}

func (w *witnessingAsCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.String("witness-code", "theCodeWeSentCertificateProvider", w.Code,
		validation.Empty(),
		validation.StringLength(4))

	return errors
}
