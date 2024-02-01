package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type organisationCreatedData struct {
	App              page.AppData
	Errors           validation.List
	OrganisationName string
}

func OrganisationCreated(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		return tmpl(w, organisationCreatedData{
			App:              appData,
			OrganisationName: organisation.Name,
		})
	}
}
