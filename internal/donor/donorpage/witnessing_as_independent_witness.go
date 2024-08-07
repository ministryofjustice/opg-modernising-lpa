package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingAsIndependentWitnessData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *witnessingAsIndependentWitnessForm
	Donor  *donordata.Provided
}

func WitnessingAsIndependentWitness(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &witnessingAsIndependentWitnessData{
			App:   appData,
			Donor: provided,
			Form:  &witnessingAsIndependentWitnessForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readWitnessingAsIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			if provided.WitnessCodeLimiter == nil {
				provided.WitnessCodeLimiter = donordata.NewLimiter(time.Minute, 5, 10)
			}

			if !provided.WitnessCodeLimiter.Allow(now()) {
				data.Errors.Add("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"})
			} else {
				code, found := provided.IndependentWitnessCodes.Find(data.Form.Code, now())
				if !found {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"})
				} else if code.HasExpired(now()) {
					data.Errors.Add("witness-code", validation.CustomError{Label: "witnessCodeExpired"})
				}
			}

			if data.Errors.None() {
				provided.WitnessCodeLimiter = nil
				provided.WitnessedByIndependentWitnessAt = now()
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if data.Errors.None() {
				return donor.PathWitnessingAsCertificateProvider.Redirect(w, r, appData, provided)
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
