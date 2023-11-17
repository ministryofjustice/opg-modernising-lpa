package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousFeeData struct {
	App     page.AppData
	Errors  validation.List
	Form    *previousFeeForm
	Options pay.PreviousFeeOptions
}

func PreviousFee(tmpl template.Template, payer Payer, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		data := &previousFeeData{
			App: appData,
			Form: &previousFeeForm{
				PreviousFee: lpa.PreviousFee,
			},
			Options: pay.PreviousFeeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousFeeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if lpa.PreviousFee != data.Form.PreviousFee {
					lpa.PreviousFee = data.Form.PreviousFee

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}
				}

				if lpa.PreviousFee.IsPreviousFeeFull() {
					return payer.Pay(appData, w, r, lpa)
				}

				return page.Paths.EvidenceRequired.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type previousFeeForm struct {
	PreviousFee pay.PreviousFee
	Error       error
}

func readPreviousFeeForm(r *http.Request) *previousFeeForm {
	previousFee, err := pay.ParsePreviousFee(page.PostFormString(r, "previous-fee"))

	return &previousFeeForm{
		PreviousFee: previousFee,
		Error:       err,
	}
}

func (f *previousFeeForm) Validate() validation.List {
	var errors validation.List

	errors.Error("previous-fee", "howMuchYouPreviouslyPaid", f.Error,
		validation.Selected())

	return errors
}
