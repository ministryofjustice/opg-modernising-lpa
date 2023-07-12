package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type removeReplacementAttorneyData struct {
	App      page.AppData
	Attorney actor.Attorney
	Errors   validation.List
	Form     *form.YesNoForm
	Options  form.YesNoOptions
}

func RemoveReplacementAttorney(logger Logger, tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		id := r.FormValue("id")
		attorney, found := lpa.ReplacementAttorneys.Get(id)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary.Format(lpa.ID))
		}

		data := &removeReplacementAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     &form.YesNoForm{},
			Options:  form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
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

				return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}
