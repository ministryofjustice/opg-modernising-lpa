package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type dashboardData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
}

func Dashboard(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &dashboardData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
