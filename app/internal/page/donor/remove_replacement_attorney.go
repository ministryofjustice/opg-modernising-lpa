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

func RemoveReplacementAttorney(logger page.Logger, tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
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

			if data.Form.RemoveAttorney == "yes" && data.Errors.None() {
				lpa.ReplacementAttorneys.Delete(attorney)
				if len(lpa.ReplacementAttorneys) == 0 {
					lpa.Tasks.ChooseReplacementAttorneys = page.TaskInProgress
				}

				err = lpaStore.Put(r.Context(), lpa)

				if err != nil {
					logger.Print(fmt.Sprintf("error removing replacement Attorney from LPA: %s", err.Error()))
					return err
				}

				if len(lpa.ReplacementAttorneys) == 0 {
					return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys)
				}

				return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
			}

			if data.Form.RemoveAttorney == "no" {
				return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
			}

		}

		return tmpl(w, data)
	}
}
