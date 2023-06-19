package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type removeAttorneyData struct {
	App      page.AppData
	Attorney actor.Attorney
	Errors   validation.List
	Form     *removeAttorneyForm
}

func RemoveAttorney(logger Logger, tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		id := r.FormValue("id")
		attorney, found := lpa.Attorneys.Get(id)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseAttorneysSummary.Format(lpa.ID))
		}

		data := &removeAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     &removeAttorneyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readRemoveAttorneyForm(r, "yesToRemoveAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.RemoveAttorney == "yes" {
					lpa.Attorneys.Delete(attorney)
					if len(lpa.Attorneys) == 1 {
						lpa.AttorneyDecisions = actor.AttorneyDecisions{}
					}

					lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
					lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						logger.Print(fmt.Sprintf("error removing Attorney from LPA: %s", err.Error()))
						return err
					}
				}

				return appData.Redirect(w, r, lpa, page.Paths.ChooseAttorneysSummary.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type removeAttorneyForm struct {
	RemoveAttorney string
	errorLabel     string
}

func readRemoveAttorneyForm(r *http.Request, errorLabel string) *removeAttorneyForm {
	return &removeAttorneyForm{
		RemoveAttorney: page.PostFormString(r, "remove-attorney"),
		errorLabel:     errorLabel,
	}
}

func (f *removeAttorneyForm) Validate() validation.List {
	var errors validation.List

	errors.String("remove-attorney", f.errorLabel, f.RemoveAttorney,
		validation.Select("yes", "no"))

	return errors
}
