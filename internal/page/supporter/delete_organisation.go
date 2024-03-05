package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteOrganisationData struct {
	App                page.AppData
	Errors             validation.List
	InProgressLPACount int
}

func DeleteOrganisation(tmpl template.Template, organisationStore OrganisationStore, sessionStore SessionStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		lpas, err := organisationStore.AllLPAs(r.Context())
		if err != nil {
			return err
		}

		data := &deleteOrganisationData{
			App:                appData,
			InProgressLPACount: len(lpas),
		}

		if r.Method == http.MethodPost {
			if err := organisationStore.SoftDelete(r.Context(), organisation); err != nil {
				return err
			}

			if err := sessionStore.ClearLogin(r, w); err != nil {
				return err
			}

			return page.Paths.Supporter.OrganisationDeleted.RedirectQuery(w, r, appData, url.Values{"organisationName": {appData.OrganisationName}})
		}

		return tmpl(w, data)
	}
}
