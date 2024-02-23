package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardData struct {
	App    page.AppData
	Errors validation.List
	Donors []actor.DonorProvidedDetails
}

func Dashboard(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		donors, err := organisationStore.AllLPAs(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &dashboardData{App: appData, Donors: donors})
	}
}
