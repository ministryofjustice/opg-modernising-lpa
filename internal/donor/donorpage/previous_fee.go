package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousFeeData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *previousFeeForm
	Options pay.PreviousFeeOptions
}

func PreviousFee(tmpl template.Template, payer Handler, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &previousFeeData{
			App: appData,
			Form: &previousFeeForm{
				PreviousFee: provided.PreviousFee,
			},
			Options: pay.PreviousFeeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousFeeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.PreviousFee != data.Form.PreviousFee {
					provided.PreviousFee = data.Form.PreviousFee

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}
				}

				if provided.PreviousFee.IsPreviousFeeFull() {
					return payer(appData, w, r, provided)
				}

				return donor.PathEvidenceRequired.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type previousFeeForm struct {
	PreviousFee pay.PreviousFee
}

func readPreviousFeeForm(r *http.Request) *previousFeeForm {
	previousFee, _ := pay.ParsePreviousFee(page.PostFormString(r, "previous-fee"))

	return &previousFeeForm{
		PreviousFee: previousFee,
	}
}

func (f *previousFeeForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("previous-fee", "howMuchYouPreviouslyPaid", f.PreviousFee,
		validation.Selected())

	return errors
}
