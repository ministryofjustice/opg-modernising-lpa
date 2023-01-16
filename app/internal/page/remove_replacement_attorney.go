package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type removeReplacementAttorneyData struct {
	App      AppData
	Attorney Attorney
	Errors   map[string]string
	Form     removeAttorneyForm
}

func RemoveReplacementAttorney(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, found := lpa.GetReplacementAttorney(id)

		if found == false {
			return appData.Lang.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
		}

		data := &removeReplacementAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     removeAttorneyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = removeAttorneyForm{
				RemoveAttorney: postFormString(r, "remove-attorney"),
			}

			data.Errors = data.Form.Validate()

			if data.Form.RemoveAttorney == "yes" && len(data.Errors) == 0 {
				lpa.DeleteReplacementAttorney(attorney)
				if len(lpa.ReplacementAttorneys) == 0 {
					lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
				}

				err = lpaStore.Put(r.Context(), appData.SessionID, lpa)

				if err != nil {
					logger.Print(fmt.Sprintf("error removing replacement Attorney from LPA: %s", err.Error()))
					return err
				}

				if len(lpa.ReplacementAttorneys) == 0 {
					return appData.Lang.Redirect(w, r, lpa, Paths.DoYouWantReplacementAttorneys)
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
			}

			if data.Form.RemoveAttorney == "no" {
				return appData.Lang.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
			}

		}

		return tmpl(w, data)
	}
}
