package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DashboardData struct {
	App    page.AppData
	Errors validation.List
}

func Dashboard(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			organisation, err := organisationStore.Get(r.Context())
			if err != nil {
				return err
			}

			donorProvided, err := organisationStore.CreateLPA(r.Context(), organisation.ID)
			if err != nil {
				return err
			}

			return page.Paths.YourDetails.Redirect(w, r, appData, donorProvided)
		}

		return tmpl(w, DashboardData{App: appData})
	}
}
