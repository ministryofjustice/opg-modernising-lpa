package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readTheLpaData struct {
	App      page.AppData
	Errors   validation.List
	Lpa      *page.Lpa
	Attorney actor.Attorney
}

func ReadTheLpa(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		attorney, ok := attorneys.Get(appData.AttorneyID)
		if !ok {
			return appData.Redirect(w, r, lpa, page.Paths.Attorney.Start)
		}

		data := &readTheLpaData{
			App:      appData,
			Lpa:      lpa,
			Attorney: attorney,
		}

		if r.Method == http.MethodPost {
			tasks := getTasks(appData, lpa)
			tasks.ReadTheLpa = page.TaskCompleted
			setTasks(appData, lpa, tasks)

			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.Attorney.RightsAndResponsibilities)
		}

		return tmpl(w, data)
	}
}
