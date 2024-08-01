package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

func PreviousFee(tmpl template.Template, payer Handler, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &previousFeeData{
			App: appData,
			Form: &previousFeeForm{
				PreviousFee: donor.PreviousFee,
			},
			Options: pay.PreviousFeeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousFeeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if donor.PreviousFee != data.Form.PreviousFee {
					donor.PreviousFee = data.Form.PreviousFee

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				if donor.PreviousFee.IsPreviousFeeFull() {
					return payer(appData, w, r, donor)
				}

				return page.Paths.EvidenceRequired.Redirect(w, r, appData, donor)
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
