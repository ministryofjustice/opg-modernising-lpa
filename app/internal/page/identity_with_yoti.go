package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type identityWithYotiData struct {
	App         AppData
	Errors      map[string]string
	ClientSdkID string
	ScenarioID  string
}

func IdentityWithYoti(tmpl template.Template, lpaStore LpaStore, yotiClient YotiClient, yotiScenarioID string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if lpa.YotiUserData.OK || yotiClient.IsTest() {
			return appData.Lang.Redirect(w, r, lpa, Paths.IdentityWithYotiCallback)
		}

		data := &identityWithYotiData{
			App:         appData,
			ClientSdkID: yotiClient.SdkID(),
			ScenarioID:  yotiScenarioID,
		}

		return tmpl(w, data)
	}
}
