package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardData struct {
	App    page.AppData
	Errors validation.List
	Lpas   []*page.Lpa
}

func Dashboard(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			switch r.PostFormValue("action") {
			case "reuse":
				lpa, err := lpaStore.Clone(r.Context(), r.PostFormValue("reuse-id"))
				if err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.YourDetails)
			default:
				lpa, err := lpaStore.Create(r.Context())
				if err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.YourDetails)
			}
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
