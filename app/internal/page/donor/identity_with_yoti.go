package donor

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

func IdentityWithYoti(tmpl template.Template, lpaStore LpaStore, sessionStore SessionStore, yotiClient YotiClient) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if lpa.DonorIdentityConfirmed() || yotiClient.IsTest() {
			return appData.Redirect(w, r, lpa, page.Paths.IdentityWithYotiCallback)
		}

		if err := sesh.SetYoti(sessionStore, r, w, &sesh.YotiSession{
			Locale: appData.Lang.String(),
			LpaID:  appData.LpaID,
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
