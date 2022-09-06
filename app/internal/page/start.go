package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type startData struct {
	App    AppData
	Errors map[string]string
}

func Start(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &startData{
			App: appData,
		}

		return tmpl(w, data)
	}
}
