package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whichFeeTypeAreYouApplyingForData struct {
	App     page.AppData
	Errors  validation.List
	Form    *whichFeeTypeAreYouApplyingForForm
	Options page.FeeTypeOptions
}

func WhichFeeTypeAreYouApplyingFor(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &whichFeeTypeAreYouApplyingForData{
			App: appData,
			Form: &whichFeeTypeAreYouApplyingForForm{
				FeeType: lpa.FeeType,
			},
			Options: page.FeeTypeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhichFeeTypeAreYouApplyingForForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.FeeType = data.Form.FeeType
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.EvidenceRequired.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type whichFeeTypeAreYouApplyingForForm struct {
	FeeType page.FeeType
	Error   error
}

func readWhichFeeTypeAreYouApplyingForForm(r *http.Request) *whichFeeTypeAreYouApplyingForForm {
	feeType, err := page.ParseFeeType(page.PostFormString(r, "fee-type"))

	return &whichFeeTypeAreYouApplyingForForm{
		FeeType: feeType,
		Error:   err,
	}
}

func (f *whichFeeTypeAreYouApplyingForForm) Validate() validation.List {
	var errors validation.List

	errors.Error("fee-type", "whichFeeTypeYouAreApplyingFor", f.Error, validation.Selected())

	return errors
}
