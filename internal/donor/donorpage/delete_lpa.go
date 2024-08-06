package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func DeleteLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if err := donorStore.Delete(r.Context()); err != nil {
				return err
			}

			return page.PathLpaDeleted.RedirectQuery(w, r, appData, url.Values{"uid": {provided.LpaUID}})
		}

		return tmpl(w, &deleteLpaData{
			App:   appData,
			Donor: provided,
		})
	}
}
