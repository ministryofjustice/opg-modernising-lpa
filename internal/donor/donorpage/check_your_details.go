package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourDetailsData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func CheckYourDetails(tmpl template.Template, shareCodeSender ShareCodeSender) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if !provided.Tasks.PayForLpa.IsCompleted() {
				return donor.PathWeHaveReceivedVoucherDetails.Redirect(w, r, appData, provided)
			}

			// TODO: MLPAB-1899 contact voucher
			if err := shareCodeSender.SendVoucherAccessCodeToDonor(r.Context(), provided, appData); err != nil {
				return err
			}

			return donor.PathWeHaveContactedVoucher.Redirect(w, r, appData, provided)
		}

		data := &checkYourDetailsData{
			App:   appData,
			Donor: provided,
		}

		return tmpl(w, data)
	}
}
