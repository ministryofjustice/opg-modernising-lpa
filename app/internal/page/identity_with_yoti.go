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

func IdentityWithYoti(tmpl template.Template, yotiClient yotiClient, yotiScenarioID string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if yotiClient.IsTest() {
			appData.Lang.Redirect(w, r, identityWithYotiCallbackPath, http.StatusFound)
			return nil
		}

		data := &identityWithYotiData{
			App:         appData,
			ClientSdkID: yotiClient.SdkID(),
			ScenarioID:  yotiScenarioID,
		}

		return tmpl(w, data)
	}
}
