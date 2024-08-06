package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whichFeeTypeAreYouApplyingForData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *whichFeeTypeAreYouApplyingForForm
	Options pay.FeeTypeOptions
}

func WhichFeeTypeAreYouApplyingFor(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whichFeeTypeAreYouApplyingForData{
			App: appData,
			Form: &whichFeeTypeAreYouApplyingForForm{
				FeeType: provided.FeeType,
			},
			Options: pay.FeeTypeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhichFeeTypeAreYouApplyingForForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.FeeType = data.Form.FeeType
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.FeeType.IsRepeatApplicationFee() {
					return donor.PathPreviousApplicationNumber.Redirect(w, r, appData, provided)
				} else {
					return donor.PathEvidenceRequired.Redirect(w, r, appData, provided)
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
