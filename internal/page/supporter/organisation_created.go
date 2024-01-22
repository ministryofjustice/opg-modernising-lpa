package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type organisationCreatedData struct {
	App              page.AppData
	Errors           validation.List
	OrganisationName string
}

func OrganisationCreated(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		organisation, err := organisationStore.Get(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, organisationCreatedData{
			App:              appData,
			OrganisationName: organisation.Name,
		})
	}
}
