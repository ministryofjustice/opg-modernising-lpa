package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouApplyingForFeeDiscountOrExemptionData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *form.YesNoForm
}

func AreYouApplyingForFeeDiscountOrExemption(tmpl template.Template, payer Handler, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &areYouApplyingForFeeDiscountOrExemptionData{
			App:                 appData,
			CertificateProvider: provided.CertificateProvider,
			Form:                form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "whetherApplyingForDifferentFeeType")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Tasks.PayForLpa = task.PaymentStateInProgress
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if data.Form.YesNo.IsNo() {
					return payer(appData, w, r, provided)
				} else {
					return donor.PathWhichFeeTypeAreYouApplyingFor.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
