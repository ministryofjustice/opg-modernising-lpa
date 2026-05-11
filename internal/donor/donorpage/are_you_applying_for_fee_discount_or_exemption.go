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

type areYouApplyingForFeeDiscountOrExemptionData struct {
	App                 appcontext.Data
	CertificateProvider donordata.CertificateProvider
	Form                *newforms.YesNoForm
}

func AreYouApplyingForFeeDiscountOrExemption(tmpl template.Template, payer Handler, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &areYouApplyingForFeeDiscountOrExemptionData{
			App:                 appData,
			CertificateProvider: provided.CertificateProvider,
			Form:                newforms.NewYesNoForm(appData.Localizer.T("whetherApplyingForDifferentFeeType")),
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				provided.Tasks.PayForLpa = task.PaymentStateInProgress
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if data.Form.YesNo.Value.IsNo() {
					if !provided.FeeType.IsFullFee() {
						provided.FeeType = pay.FullFee
						if err := donorStore.Put(r.Context(), provided); err != nil {
							return err
						}
					}
					return payer(appData, w, r, provided)
				} else {
					return donor.PathWhichFeeTypeAreYouApplyingFor.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
