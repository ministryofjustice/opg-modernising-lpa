package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeAttorneyData struct {
	App        page.AppData
	TitleLabel string
	Name       string
	Errors     validation.List
	Form       *form.YesNoForm
}

func RemoveAttorney(logger Logger, tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		attorney, found := donor.Attorneys.Get(actor.UIDFromRequest(r))

		if found == false {
			return page.Paths.ChooseAttorneysSummary.Redirect(w, r, appData, donor)
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
					donor.Attorneys.Delete(attorney)
					if donor.Attorneys.Len() == 1 {
						donor.AttorneyDecisions = actor.AttorneyDecisions{}
					}

					donor.Tasks.ChooseAttorneys = page.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
					donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

					if err := donorStore.Put(r.Context(), donor); err != nil {
						logger.Print(fmt.Sprintf("error removing Attorney from LPA: %s", err.Error()))
						return err
					}
				}

				return page.Paths.ChooseAttorneysSummary.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
