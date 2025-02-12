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

func WithdrawLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time, lpaStoreClient LpaStoreClient, notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			provided.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if err := lpaStoreClient.SendDonorWithdrawLPA(r.Context(), provided.LpaUID); err != nil {
				return err
			}

			if !provided.CertificateProviderInvitedAt.IsZero() {
				email := notify.InformCertificateProviderLPAHasBeenRevoked{
					DonorFullName:                   provided.Donor.FullName(),
					DonorFullNamePossessive:         appData.Localizer.Possessive(provided.Donor.FullName()),
					LpaType:                         localize.LowerFirst(appData.Localizer.T(provided.Type.String())),
					CertificateProviderFullName:     provided.CertificateProvider.FullName(),
					InvitedDate:                     appData.Localizer.FormatDate(provided.CertificateProviderInvitedAt),
					CertificateProviderStartPageURL: appPublicURL + appData.Lang.URL(page.PathCertificateProviderStart.Format()),
				}

				if err := notifyClient.SendActorEmail(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), provided.LpaUID, email); err != nil {
					return fmt.Errorf("error sending LPA revoked email to certificate provider: %v", err)
				}
			}

			return page.PathLpaWithdrawn.RedirectQuery(w, r, appData, url.Values{"uid": {provided.LpaUID}})
		}

		return tmpl(w, &withdrawLpaData{
			App:   appData,
			Donor: provided,
		})
	}
}
