package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsIndependentWitnessData struct {
	App    page.AppData
	Errors validation.List
	Form   *witnessingAsIndependentWitnessForm
	Donor  *actor.DonorProvidedDetails
}

func WitnessingAsIndependentWitness(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &witnessingAsIndependentWitnessData{
			App:   appData,
			Donor: donor,
			Form:  &witnessingAsIndependentWitnessForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			if donor.WitnessCodeLimiter == nil {
				donor.WitnessCodeLimiter = actor.NewLimiter(time.Minute, 5, 10)
			}

			if !donor.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := donor.IndependentWitnessCodes.Find(data.Form.Code)
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired() {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				donor.WitnessCodeLimiter = nil
				donor.WitnessedByIndependentWitnessAt = now()
			}

			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			if data.Errors.None() {
				return page.Paths.WitnessingAsCertificateProvider.Redirect(w, r, appData, donor)
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
