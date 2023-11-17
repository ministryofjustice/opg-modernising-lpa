package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsIndependentWitnessData struct {
	App    page.AppData
	Errors validation.List
	Form   *witnessingAsIndependentWitnessForm
	Lpa    *page.Lpa
}

func WitnessingAsIndependentWitness(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &witnessingAsIndependentWitnessData{
			App:  appData,
			Lpa:  lpa,
			Form: &witnessingAsIndependentWitnessForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			if lpa.WitnessCodeLimiter == nil {
				lpa.WitnessCodeLimiter = page.NewLimiter(time.Minute, 5, 10)
			}

			if !lpa.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := lpa.IndependentWitnessCodes.Find(data.Form.Code)
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired() {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				lpa.WitnessCodeLimiter = nil
				lpa.WitnessedByIndependentWitnessAt = now()
			}

			if err := donorStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if data.Errors.None() {
				return page.Paths.WitnessingAsCertificateProvider.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type witnessingAsIndependentWitnessForm struct {
	Code string
}

func readWitnessingAsIndependentWitnessForm(r *http.Request) *witnessingAsIndependentWitnessForm {
	return &witnessingAsIndependentWitnessForm{
		Code: page.PostFormString(r, "witness-code"),
	}
}

func (w *witnessingAsIndependentWitnessForm) Validate() validation.List {
	var errors validation.List

	errors.String("witness-code", "theCodeWeSentIndependentWitness", w.Code,
		validation.Empty(),
		validation.StringLength(4))

	return errors
}
