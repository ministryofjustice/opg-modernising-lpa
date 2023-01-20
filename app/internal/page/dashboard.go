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
		if r.Method == http.MethodPost {
			lpa, err := lpaStore.Create(r.Context())
			if err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, Paths.YourDetails)
		}

		lpa, err := lpaStore.GetAll(r.Context())
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
