package donorpage

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func DeleteLpa(logger Logger, tmpl template.Template, donorStore DonorStore, notifyClient NotifyClient, certificateProviderStartURL string, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if !provided.CertificateProviderInvitedAt.IsZero() {
				email := notify.InformCertificateProviderLPAHasBeenDeleted{
					DonorFullName:                   provided.Donor.FullName(),
					DonorFullNamePossessive:         appData.Localizer.Possessive(provided.Donor.FullName()),
					LpaType:                         localize.LowerFirst(appData.Localizer.T(provided.Type.String())),
					CertificateProviderFullName:     provided.CertificateProvider.FullName(),
					InvitedDate:                     appData.Localizer.FormatDate(provided.CertificateProviderInvitedAt),
					CertificateProviderStartPageURL: certificateProviderStartURL,
				}

				if err := notifyClient.SendActorEmail(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), provided.LpaUID, email); err != nil {
					return fmt.Errorf("error sending LPA deleted email to certificate provider: %w", err)
				}
			}

			if !provided.VoucherInvitedAt.IsZero() {
				if err := notifyClient.SendActorEmail(r.Context(), notify.ToVoucher(provided.Voucher), provided.LpaUID, notify.VoucherLpaDeleted{
					DonorFullName:           provided.Donor.FullName(),
					DonorFullNamePossessive: appData.Localizer.Possessive(provided.Donor.FullName()),
					InvitedDate:             appData.Localizer.FormatDate(provided.VoucherInvitedAt),
					LpaType:                 localize.LowerFirst(appData.Localizer.T(provided.Type.String())),
					VoucherFullName:         provided.Voucher.FullName(),
				}); err != nil {
					return fmt.Errorf("error sending voucher email: %w", err)
				}
			}

			if err := donorStore.Delete(r.Context()); err != nil {
				return fmt.Errorf("error deleting lpa: %w", err)
			}

			if err := eventClient.SendMetric(r.Context(), provided.LpaID, event.CategoryDraftLPADeleted, event.MeasureOnlineDonor); err != nil {
				logger.ErrorContext(r.Context(), "error sending metric", slog.Any("err", err))
			}

			return page.PathLpaDeleted.RedirectQuery(w, r, appData, url.Values{"uid": {provided.LpaUID}})
		}

		return tmpl(w, &deleteLpaData{
			App:   appData,
			Donor: provided,
		})
	}
}
