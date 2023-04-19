package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithYotiData struct {
	App         page.AppData
	Errors      validation.List
	ClientSdkID string
	ScenarioID  string
}

func IdentityWithYoti(tmpl template.Template, sessionStore SessionStore, yotiClient YotiClient, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if certificateProvider.CertificateProviderIdentityConfirmed() || yotiClient.IsTest() {
			return appData.Redirect(w, r, nil, page.Paths.CertificateProviderIdentityWithYotiCallback)
		}

		if err := sesh.SetYoti(sessionStore, r, w, &sesh.YotiSession{
			Locale:              appData.Lang.String(),
			LpaID:               certificateProvider.LpaID,
			CertificateProvider: true,
		}); err != nil {
			return err
		}

		data := &identityWithYotiData{
			App:         appData,
			ClientSdkID: yotiClient.SdkID(),
			ScenarioID:  yotiClient.ScenarioID(),
		}

		return tmpl(w, data)
	}
}
