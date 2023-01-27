package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardData struct {
	App    AppData
	Errors validation.List
	Lpas   []*Lpa
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

		lpas, err := lpaStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		data := &dashboardData{
			App:  appData,
			Lpas: lpas,
		}

		return tmpl(w, data)
	}
}
