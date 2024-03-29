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
	Donor   *actor.DonorProvidedDetails
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    donor.ReplacementAttorneyDecisions.How,
				DecisionsDetails: donor.ReplacementAttorneyDecisions.Details,
			},
			Options: actor.AttorneysActValues,
			Donor:   donor,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howReplacementAttorneysShouldMakeDecisions", "detailsAboutTheDecisionsYourReplacementAttorneysMustMakeTogether")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.ReplacementAttorneyDecisions = actor.MakeAttorneyDecisions(
					donor.ReplacementAttorneyDecisions,
					data.Form.DecisionsType,
					data.Form.DecisionsDetails)
				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
