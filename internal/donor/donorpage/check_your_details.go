package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourDetailsData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func CheckYourDetails(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if !donor.Tasks.PayForLpa.IsCompleted() {
				return page.Paths.WeHaveReceivedVoucherDetails.Redirect(w, r, appData, donor)
			}

			// TODO: MLPAB-1897 send code to donor and MLPAB-1899 contact voucher

			return page.Paths.WeHaveContactedVoucher.Redirect(w, r, appData, donor)
		}

		data := &checkYourDetailsData{
			App:   appData,
			Donor: donor,
		}

		return tmpl(w, data)
	}
}