package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type costOfRepeatApplicationData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.SelectForm[pay.CostOfRepeatApplication, pay.CostOfRepeatApplicationOptions, *pay.CostOfRepeatApplication]
}

func CostOfRepeatApplication(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &costOfRepeatApplicationData{
			App:  appData,
			Form: form.NewSelectForm(provided.CostOfRepeatApplication, pay.CostOfRepeatApplicationValues, "whichFeeYouAreEligibleToPay"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.CostOfRepeatApplication != data.Form.Selected {
					provided.CostOfRepeatApplication = data.Form.Selected
					provided.Tasks.PayForLpa = task.PaymentStatePending

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}
				}

				if provided.CostOfRepeatApplication.IsHalfFee() {
					return donor.PathPreviousFee.Redirect(w, r, appData, provided)
				}

				return donor.PathWhatHappensNextRepeatApplicationNoFee.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
