package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type identityWithTodoData struct {
	App            AppData
	Errors         map[string]string
	IdentityOption IdentityOption
}

func IdentityWithTodo(tmpl template.Template, dataStore DataStore, identityOption IdentityOption) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			var lpa Lpa
			if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
				return err
			}

			appData.Lang.Redirect(w, r, lpa.IdentityOptions.NextPath(identityOption), http.StatusFound)
			return nil
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
