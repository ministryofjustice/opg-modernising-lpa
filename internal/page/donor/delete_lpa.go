package donor

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *actor.DonorProvidedDetails
}

func DeleteLpa(tmpl template.Template, donorStore DonorStore) Handler {

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		if r.Method == http.MethodPost {
			if err := donorStore.Delete(r.Context()); err != nil {
				return err
			}

			return page.Paths.LpaDeleted.RedirectQuery(w, r, appData, url.Values{"uid": {lpa.UID}})
		}

		return tmpl(w, &deleteLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
