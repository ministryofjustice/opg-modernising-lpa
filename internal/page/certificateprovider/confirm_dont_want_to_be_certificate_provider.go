package certificateprovider

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ConfirmDontWantToBeCertificateProvider(tmpl template.Template, shareCodeStore ShareCodeStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: r.URL.Query().Get("LpaID")})

		lpa, err := lpaStoreResolvingService.Get(ctx)
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeCertificateProviderData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			if !lpa.SignedAt.IsZero() {
				if err := lpaStoreClient.SendCertificateProviderOptOut(ctx, lpa.LpaUID); err != nil {
					return err
				}
			}

			if referenceNumber := r.URL.Query().Get("referenceNumber"); referenceNumber != "" {
				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, referenceNumber)
				if err != nil {
					return err
				}

				if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
					return err
				}
			}

			return page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}
