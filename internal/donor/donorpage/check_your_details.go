package donorpage

import (
	"net/http"
	"time"

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

func CheckYourDetails(tmpl template.Template, shareCodeSender ShareCodeSender, now func() time.Time, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if provided.Tasks.PayForLpa.IsCompleted() {
				if err := shareCodeSender.SendVoucherAccessCode(r.Context(), provided, appData); err != nil {
					return err
				}

				provided.VoucherInvitedAt = now()

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}
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
