package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeReplacementAttorneyData struct {
	App      AppData
	Attorney Attorney
	Errors   validation.List
	Form     *removeAttorneyForm
}

func RemoveReplacementAttorney(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, found := lpa.GetReplacementAttorney(id)

		if found == false {
			return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
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
				lpa.DeleteReplacementAttorney(attorney)
				if len(lpa.ReplacementAttorneys) == 0 {
					lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
				}

				err = lpaStore.Put(r.Context(), lpa)

				if err != nil {
					logger.Print(fmt.Sprintf("error removing replacement Attorney from LPA: %s", err.Error()))
					return err
				}

				if len(lpa.ReplacementAttorneys) == 0 {
					return appData.Redirect(w, r, lpa, Paths.DoYouWantReplacementAttorneys)
				}

				return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
			}

			if data.Form.RemoveAttorney == "no" {
				return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
			}

		}

		return tmpl(w, data)
	}
}
