package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func RemoveReplacementAttorney(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		attorney, found := donor.ReplacementAttorneys.Get(actoruid.FromRequest(r))

		if found == false {
			return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, donor)
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
					donor.ReplacementAttorneys.Delete(attorney)
					if donor.ReplacementAttorneys.Len() == 1 {
						donor.ReplacementAttorneyDecisions = donordata.AttorneyDecisions{}
					}

					donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return fmt.Errorf("error removing replacement Attorney from LPA: %w", err)
					}
				}

				return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
