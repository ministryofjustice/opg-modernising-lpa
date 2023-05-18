package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeReplacementAttorneyData struct {
	App      page.AppData
	Attorney actor.Attorney
	Errors   validation.List
	Form     *removeAttorneyForm
}

func RemoveReplacementAttorney(logger Logger, tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, found := lpa.ReplacementAttorneys.Get(id)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
		}

		data := &removeReplacementAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     &removeAttorneyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readRemoveAttorneyForm(r, "yesToRemoveReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.RemoveAttorney == "yes" {
					lpa.ReplacementAttorneys.Delete(attorney)
					if len(lpa.ReplacementAttorneys) == 1 {
						lpa.ReplacementAttorneyDecisions = actor.AttorneyDecisions{}
					}

					lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						logger.Print(fmt.Sprintf("error removing replacement Attorney from LPA: %s", err.Error()))
						return err
					}
				}

				return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
			}
		}

		return tmpl(w, data)
	}
}
