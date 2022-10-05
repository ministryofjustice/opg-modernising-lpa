package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type identityWithEasyIDCallbackData struct {
	App      AppData
	Errors   map[string]string
	FullName string
}

func IdentityWithEasyIDCallback(tmpl template.Template, yotiClient yotiClient) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			appData.Lang.Redirect(w, r, identityOptionRedirectPath, http.StatusFound)
			return nil
		}

		user, err := yotiClient.User(r.FormValue("token"))
		if err != nil {
			return err
		}

		data := &identityWithEasyIDCallbackData{
			App:      appData,
			FullName: user.FullName,
		}

		return tmpl(w, data)
	}
}
