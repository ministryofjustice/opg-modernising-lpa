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

func IdentityWithTodo(tmpl template.Template, identityOption IdentityOption) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			return appData.Lang.Redirect(w, r, nil, Paths.ReadYourLpa)
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
