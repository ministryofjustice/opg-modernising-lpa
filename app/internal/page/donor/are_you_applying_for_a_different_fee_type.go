package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type areYouApplyingForADifferentFeeTypeData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Options             form.YesNoOptions
	Form                *form.YesNoForm
}

func AreYouApplyingForADifferentFeeType(tmpl template.Template, payer Payer) Handler {
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
