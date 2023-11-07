package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App     page.AppData
	Errors  validation.List
	Form    *howShouldAttorneysMakeDecisionsForm
	Options actor.AttorneysActOptions
	Lpa     *page.Lpa
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.ReplacementAttorneyDecisions.How,
				DecisionsDetails: lpa.ReplacementAttorneyDecisions.Details,
			},
			Options: actor.AttorneysActValues,
			Lpa:     lpa,
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

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}
