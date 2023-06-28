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
	Options  actor.YesNoOptions
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
			Options:  actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readRemoveAttorneyForm(r, "yesToRemoveAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.RemoveAttorney == actor.Yes {
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
	RemoveAttorney actor.YesNo
	Error          error
	errorLabel     string
}

func readRemoveAttorneyForm(r *http.Request, errorLabel string) *removeAttorneyForm {
	remove, err := actor.ParseYesNo(page.PostFormString(r, "remove-attorney"))

	return &removeAttorneyForm{
		RemoveAttorney: remove,
		Error:          err,
		errorLabel:     errorLabel,
	}
}

func (f *removeAttorneyForm) Validate() validation.List {
	var errors validation.List

	errors.Error("remove-attorney", f.errorLabel, f.Error,
		validation.Selected())

	return errors
}
