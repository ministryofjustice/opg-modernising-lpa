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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeAttorneyData struct {
	App        appcontext.Data
	TitleLabel string
	Name       string
	Errors     validation.List
	Form       *form.YesNoForm
}

func RemoveAttorney(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorney, found := provided.Attorneys.Get(actoruid.FromRequest(r))

		if found == false {
			return donor.PathChooseAttorneysSummary.Redirect(w, r, appData, provided)
		}

		data := &removeAttorneyData{
			App:        appData,
			TitleLabel: "removeAnAttorney",
			Name:       attorney.FullName(),
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					provided.Attorneys.Delete(attorney)
					if provided.Attorneys.Len() == 1 {
						provided.AttorneyDecisions = donordata.AttorneyDecisions{}
					}

					provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
					provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return fmt.Errorf("error removing Attorney from LPA: %w", err)
					}
				}

				return donor.PathChooseAttorneysSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
