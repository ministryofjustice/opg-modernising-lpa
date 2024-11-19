package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousFeeData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.SelectForm[pay.PreviousFee, pay.PreviousFeeOptions, *pay.PreviousFee]
}

func PreviousFee(tmpl template.Template, payer Handler, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &previousFeeData{
			App:  appData,
			Form: form.NewSelectForm(provided.PreviousFee, pay.PreviousFeeValues, "howMuchYouPreviouslyPaid"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.PreviousFee != data.Form.Selected {
					provided.PreviousFee = data.Form.Selected

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
