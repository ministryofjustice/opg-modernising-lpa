package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type readTheLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func ReadTheLpa(tmpl template.Template, donorStore DonorStore, attorneyStore AttorneyStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		if r.Method == http.MethodPost {
			attorneyProvidedDetails.Tasks.ReadTheLpa = actor.TaskCompleted

			if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
				return err
			}

			return appData.Redirect(w, r, nil, page.Paths.Attorney.RightsAndResponsibilities.Format(attorneyProvidedDetails.LpaID))
		}

		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		data := &readTheLpaData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
