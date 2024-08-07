package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func RemoveReplacementAttorney(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorney, found := provided.ReplacementAttorneys.Get(actoruid.FromRequest(r))

		if found == false {
			return donor.PathChooseReplacementAttorneysSummary.Redirect(w, r, appData, provided)
		}

		data := &removeAttorneyData{
			App:        appData,
			TitleLabel: "doYouWantToRemoveReplacementAttorney",
			Name:       attorney.FullName(),
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					provided.ReplacementAttorneys.Delete(attorney)
					if provided.ReplacementAttorneys.Len() == 1 {
						provided.ReplacementAttorneyDecisions = donordata.AttorneyDecisions{}
					}

					provided.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(provided)

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return fmt.Errorf("error removing replacement Attorney from LPA: %w", err)
					}
				}

				return donor.PathChooseReplacementAttorneysSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
