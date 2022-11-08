package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseReplacementAttorneysData struct {
	App    AppData
	Errors map[string]string
}

func ChooseReplacementAttorneys(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &chooseReplacementAttorneysData{
			App: appData,
		}
		return tmpl(w, data)
	}
}
