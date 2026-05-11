package supporterpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

type editOrganisationNameData struct {
	App  appcontext.Data
	Form *organisationNameForm
}

func EditOrganisationName(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		data := &editOrganisationNameData{
			App:  appData,
			Form: newOrganisationNameForm(appData.Localizer.T("yourOrganisationName")),
		}

		data.Form.Name.SetInput(organisation.Name)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			organisation.Name = data.Form.Name.Value

			if err := organisationStore.Put(r.Context(), organisation); err != nil {
				return err
			}

			return supporter.PathOrganisationDetails.RedirectQuery(w, r, appData, url.Values{"updated": {"name"}})
		}

		return tmpl(w, data)
	}
}
