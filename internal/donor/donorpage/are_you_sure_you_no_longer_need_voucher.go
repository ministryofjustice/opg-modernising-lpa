package donorpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouSureYouNoLongerNeedVoucherData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func AreYouSureYouNoLongerNeedVoucher(tmpl template.Template, donorStore DonorStore, notifyClient NotifyClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &areYouSureYouNoLongerNeedVoucherData{
			App:   appData,
			Donor: provided,
		}

		if r.Method == http.MethodPost {
			voucherFullName := provided.Voucher.FullName()

			doNext, err := donordata.ParseNoVoucherDecision(r.FormValue("choice"))
			if err != nil {
				return err
			}

			provided.WantVoucher = form.YesNoUnknown
			nextPage := handleDoNext(doNext, provided).Format(provided.LpaID)

			if err := notifyClient.SendActorEmail(r.Context(), localize.En, provided.Voucher.Email, provided.LpaUID, notify.VoucherInformedTheyAreNoLongerNeededToVouchEmail{
				VoucherFullName: provided.Voucher.FullName(),
				DonorFullName:   provided.Donor.FullName(),
			}); err != nil {
				return fmt.Errorf("failed to send email: %w", err)
			}

			if err := donorStore.DeleteVoucher(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathWeHaveInformedVoucherNoLongerNeeded.RedirectQuery(w, r, appData, provided, url.Values{
				"choice":          {r.FormValue("choice")},
				"next":            {nextPage},
				"voucherFullName": {voucherFullName},
			})
		}

		return tmpl(w, data)
	}
}
