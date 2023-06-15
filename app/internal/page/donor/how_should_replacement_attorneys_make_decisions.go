package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App    page.AppData
	Errors validation.List
	Form   *howShouldAttorneysMakeDecisionsForm
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.ReplacementAttorneyDecisions.How,
				DecisionsDetails: lpa.ReplacementAttorneyDecisions.Details,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howReplacementAttorneysShouldMakeDecisions")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.ReplacementAttorneyDecisions = actor.MakeAttorneyDecisions(
					lpa.ReplacementAttorneyDecisions,
					data.Form.DecisionsType,
					data.Form.DecisionsDetails)
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.ReplacementAttorneyDecisions.RequiresHappiness(len(lpa.ReplacementAttorneys)) {
					return appData.Redirect(w, r, lpa, page.Paths.AreYouHappyIfOneReplacementAttorneyCantActNoneCan)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.TaskList)
				}
			}
		}

		return tmpl(w, data)
	}
}
