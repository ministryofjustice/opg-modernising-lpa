package certificateproviderpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

func EnterAccessCode(sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreClient LpaStoreClient, dashboardStore DashboardStore, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, shareCode sharecodedata.Link) error {
		if lpa.CertificateProvider.Channel.IsPaper() && !lpa.CertificateProvider.SignedAt.IsZero() {
			if err := lpaStoreClient.SendPaperCertificateProviderAccessOnline(r.Context(), lpa, session.Email); err != nil {
				return fmt.Errorf("sending certificate provider email to LPA store: %w", err)
			}

			redirectTo := page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn

			results, err := dashboardStore.GetAll(r.Context())
			if err != nil {
				return fmt.Errorf("getting dashboard results: %w", err)
			}

			if results.Empty() {
				if err = sessionStore.ClearLogin(r, w); err != nil {
					return fmt.Errorf("clearing login session: %w", err)
				}

				redirectTo = page.PathCertificateProviderYouHaveAlreadyProvidedACertificate
			}

			return redirectTo.RedirectQuery(w, r, appData, url.Values{
				"donorFullName": {lpa.Donor.FullName()},
				"lpaType":       {lpa.Type.String()},
			})
		}

		if _, err := certificateProviderStore.Create(r.Context(), shareCode, session.Email); err != nil {
			return fmt.Errorf("creating certificate provider: %w", err)
		}

		if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineCertificateProvider); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return certificateprovider.PathWhoIsEligible.Redirect(w, r, appData, appData.LpaID)
	}
}
