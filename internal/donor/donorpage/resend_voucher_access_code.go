package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type resendVoucherAccessCodeData struct {
	App    appcontext.Data
	Errors validation.List
}

func ResendVoucherAccessCode(tmpl template.Template, accessCodeSender AccessCodeSender) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &resendVoucherAccessCodeData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			if err := accessCodeSender.SendVoucherAccessCode(r.Context(), provided, appData); err != nil {
				return err
			}

			return donor.PathWeHaveContactedVoucher.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
