package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type identityWithEasyIDData struct {
	App         AppData
	Errors      map[string]string
	ClientSdkID string
	ScenarioID  string
}

func IdentityWithEasyID(tmpl template.Template, yotiClient yotiClient, yotiScenarioID string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if yotiClient.IsTest() {
			appData.Lang.Redirect(w, r, identityWithEasyIDCallbackPath, http.StatusFound)
			return nil
		}

		data := &identityWithEasyIDData{
			App:         appData,
			ClientSdkID: yotiClient.SdkID(),
			ScenarioID:  yotiScenarioID,
		}

		return tmpl(w, data)
	}
}
