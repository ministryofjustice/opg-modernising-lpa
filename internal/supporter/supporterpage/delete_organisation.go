package supporterpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteOrganisationData struct {
	App                appcontext.Data
	Errors             validation.List
	InProgressLPACount int
}

func DeleteOrganisation(logger Logger, tmpl template.Template, organisationStore OrganisationStore, sessionStore SessionStore, searchClient SearchClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		if r.Method == http.MethodPost {
			if err := organisationStore.SoftDelete(r.Context(), organisation); err != nil {
				return err
			}
			logger.InfoContext(r.Context(), "organisation deleted")

			if err := sessionStore.ClearLogin(r, w); err != nil {
				return err
			}

			return page.Paths.Supporter.OrganisationDeleted.RedirectQuery(w, r, appData, url.Values{"organisationName": {appData.SupporterData.OrganisationName}})
		}

		inProgressLPACount, err := searchClient.CountWithQuery(r.Context(), search.CountWithQueryReq{MustNotExist: "RegisteredAt"})

		if err != nil {
			return err
		}

		data := &deleteOrganisationData{
			App:                appData,
			InProgressLPACount: inProgressLPACount,
		}

		return tmpl(w, data)
	}
}
