package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type costOfRepeatApplicationData struct {
	App  appcontext.Data
	Form *newforms.EnumForm[pay.CostOfRepeatApplication, pay.CostOfRepeatApplicationOptions, *pay.CostOfRepeatApplication]
}

func CostOfRepeatApplication(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &costOfRepeatApplicationData{
			App:  appData,
			Form: newforms.NewEnumForm[pay.CostOfRepeatApplication](appData.Localizer.T("whichFeeYouAreEligibleToPay"), pay.CostOfRepeatApplicationValues),
		}

		data.Form.Enum.SetInput(provided.CostOfRepeatApplication)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if provided.CostOfRepeatApplication != data.Form.Enum.Value {
				provided.CostOfRepeatApplication = data.Form.Enum.Value
				if provided.CostOfRepeatApplication.IsNoFee() {
					provided.Tasks.PayForLpa = task.PaymentStatePending
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}
			}

			if provided.CostOfRepeatApplication.IsNoFee() {
				return donor.PathWhatHappensNextRepeatApplicationNoFee.Redirect(w, r, appData, provided)
			}

			return donor.PathPreviousFee.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
