package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readTheLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ReadTheLpa(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
		if r.Method == http.MethodPost {
			attorneyProvidedDetails.Tasks.ReadTheLpa = task.StateCompleted

			if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
				return err
			}

			return attorney.PathTaskList.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		data := &readTheLpaData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
