package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DashboardData struct {
	App    page.AppData
	Errors validation.List
}

func Dashboard(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		if r.Method == http.MethodPost {
			donorProvided, err := organisationStore.CreateLPA(r.Context(), organisation.ID)
			if err != nil {
				return err
			}

			return page.Paths.YourDetails.Redirect(w, r.WithContext(r.Context()), appData, donorProvided)
		}

		return tmpl(w, DashboardData{App: appData})
	}
}
