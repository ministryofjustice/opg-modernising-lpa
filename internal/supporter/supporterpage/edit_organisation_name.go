package supporterpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type editOrganisationNameData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *organisationNameForm
}

func EditOrganisationName(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		data := &editOrganisationNameData{
			App: appData,
			Form: &organisationNameForm{
				Name: organisation.Name,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readOrganisationNameForm(r, "yourOrganisationName")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				organisation.Name = data.Form.Name
				if err := organisationStore.Put(r.Context(), organisation); err != nil {
					return err
				}

				return page.Paths.Supporter.OrganisationDetails.RedirectQuery(w, r, appData, url.Values{"updated": {"name"}})
			}
		}

		return tmpl(w, data)
	}
}
