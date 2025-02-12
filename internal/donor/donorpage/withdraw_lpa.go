package donorpage

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type withdrawLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func WithdrawLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time, lpaStoreClient LpaStoreClient, notifyClient NotifyClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if !provided.VoucherInvitedAt.IsZero() {
				if err := notifyClient.SendActorEmail(r.Context(), notify.ToVoucher(provided.Voucher), provided.LpaUID, notify.VoucherLpaRevoked{
					DonorFullName:           provided.Donor.FullName(),
					DonorFullNamePossessive: appData.Localizer.Possessive(provided.Donor.FullName()),
					InvitedDate:             appData.Localizer.FormatDate(provided.VoucherInvitedAt),
					LpaType:                 localize.LowerFirst(appData.Localizer.T(provided.Type.String())),
					VoucherFullName:         provided.Voucher.FullName(),
				}); err != nil {
					return fmt.Errorf("error sending voucher email: %w", err)
				}
			}

			provided.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if err := lpaStoreClient.SendDonorWithdrawLPA(r.Context(), provided.LpaUID); err != nil {
				return err
			}

			return page.PathLpaWithdrawn.RedirectQuery(w, r, appData, url.Values{"uid": {provided.LpaUID}})
		}

		return tmpl(w, &withdrawLpaData{
			App:   appData,
			Donor: provided,
		})
	}
}
