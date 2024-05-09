package certificateprovider

import (
	"net/http"

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
		shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, r.URL.Query().Get("referenceNumber"))
		if err != nil {
			return err
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: shareCode.LpaOwnerKey.SK(), LpaID: shareCode.LpaKey.ID()})

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

			if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.Redirect(w, r, appData)
		}

		return tmpl(w, data)
	}
}
