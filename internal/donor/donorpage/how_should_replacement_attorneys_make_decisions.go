package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *howShouldAttorneysMakeDecisionsForm
	Options lpadata.AttorneysActOptions
	Donor   *donordata.Provided
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    provided.ReplacementAttorneyDecisions.How,
				DecisionsDetails: provided.ReplacementAttorneyDecisions.Details,
			},
			Options: lpadata.AttorneysActValues,
			Donor:   provided,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howReplacementAttorneysShouldMakeDecisions", "detailsAboutTheDecisionsYourReplacementAttorneysMustMakeJointly")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.ReplacementAttorneyDecisions = donordata.MakeAttorneyDecisions(
					provided.ReplacementAttorneyDecisions,
					data.Form.DecisionsType,
					data.Form.DecisionsDetails)
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
