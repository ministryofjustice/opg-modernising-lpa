package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouApplyingForFeeDiscountOrExemptionData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Form                *form.YesNoForm
}

func AreYouApplyingForFeeDiscountOrExemption(tmpl template.Template, payer Payer, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &areYouApplyingForFeeDiscountOrExemptionData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			Form:                form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "whetherApplyingForDifferentFeeType")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Tasks.PayForLpa = actor.PaymentTaskInProgress
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if data.Form.YesNo.IsNo() {
					return payer.Pay(appData, w, r, donor)
				} else {
					return page.Paths.WhichFeeTypeAreYouApplyingFor.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}
