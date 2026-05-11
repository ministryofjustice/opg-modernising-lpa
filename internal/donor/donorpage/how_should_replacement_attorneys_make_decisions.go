package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App   appcontext.Data
	Form  *howShouldAttorneysMakeDecisionsForm
	Donor *donordata.Provided
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App:   appData,
			Form:  newHowShouldAttorneysMakeDecisionsForm(appData.Localizer, "howReplacementAttorneysShouldMakeDecisions", "detailsAboutTheDecisionsYourReplacementAttorneysMustMakeJointly"),
			Donor: provided,
		}

		data.Form.DecisionsType.SetInput(provided.ReplacementAttorneyDecisions.How)
		data.Form.DecisionsDetails.SetInput(provided.ReplacementAttorneyDecisions.Details)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			provided.ReplacementAttorneyDecisions.How = data.Form.DecisionsType.Value
			provided.ReplacementAttorneyDecisions.Details = data.Form.DecisionsDetails.Value
			provided.UpdateDecisions()
			provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
