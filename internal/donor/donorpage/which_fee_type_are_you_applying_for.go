package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whichFeeTypeAreYouApplyingForData struct {
	App     page.AppData
	Errors  validation.List
	Form    *whichFeeTypeAreYouApplyingForForm
	Options pay.FeeTypeOptions
}

func WhichFeeTypeAreYouApplyingFor(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &whichFeeTypeAreYouApplyingForData{
			App: appData,
			Form: &whichFeeTypeAreYouApplyingForForm{
				FeeType: donor.FeeType,
			},
			Options: pay.FeeTypeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhichFeeTypeAreYouApplyingForForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.FeeType = data.Form.FeeType
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.FeeType.IsRepeatApplicationFee() {
					return page.Paths.PreviousApplicationNumber.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.EvidenceRequired.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}

type whichFeeTypeAreYouApplyingForForm struct {
	FeeType pay.FeeType
	Error   error
}

func readWhichFeeTypeAreYouApplyingForForm(r *http.Request) *whichFeeTypeAreYouApplyingForForm {
	feeType, err := pay.ParseFeeType(page.PostFormString(r, "fee-type"))

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
