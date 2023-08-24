package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouApplyingForADifferentFeeTypeData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Options             form.YesNoOptions
	Form                *form.YesNoForm
}

func AreYouApplyingForADifferentFeeType(tmpl template.Template, payer Payer, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &areYouApplyingForADifferentFeeTypeData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Options:             form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "whetherApplyingForDifferentFeeType")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.Tasks.PayForLpa = actor.PaymentTaskInProgress
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if data.Form.YesNo.IsNo() {
					return payer.Pay(appData, w, r, lpa)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.WhichFeeTypeAreYouApplyingFor.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}
