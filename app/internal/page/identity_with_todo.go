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

func IdentityWithTodo(tmpl template.Template, lpaStore LpaStore, identityOption IdentityOption) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
			if err != nil {
				return err
			}

			appData.Lang.Redirect(w, r, lpa.IdentityOptions.NextPath(identityOption, appData.Paths), http.StatusFound)
			return nil
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
