package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type startData struct {
	App AppData
}

func Start(logger Logger, tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) {
		data := &startData{
			App: appData,
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}
