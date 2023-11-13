package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func RemoveTrustCorporation(tmpl template.Template, donorStore DonorStore, isReplacement bool) Handler {
	redirect := page.Paths.ChooseAttorneysSummary
	if isReplacement {
		redirect = page.Paths.ChooseReplacementAttorneysSummary
	}

	updateLpa := func(lpa *page.Lpa) {
		lpa.Attorneys.TrustCorporation = actor.TrustCorporation{}
		if lpa.Attorneys.Len() == 1 {
			lpa.AttorneyDecisions = actor.AttorneyDecisions{}
		}
	}

	if isReplacement {
		updateLpa = func(lpa *page.Lpa) {
			lpa.ReplacementAttorneys.TrustCorporation = actor.TrustCorporation{}
			if lpa.ReplacementAttorneys.Len() == 1 {
				lpa.ReplacementAttorneyDecisions = actor.AttorneyDecisions{}
			}
		}
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		name := lpa.Attorneys.TrustCorporation.Name
		if isReplacement {
			name = lpa.ReplacementAttorneys.TrustCorporation.Name
		}

		data := &removeAttorneyData{
			App:     appData,
			Name:    name,
			Form:    &form.YesNoForm{},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveTrustCorporation")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo.IsYes() {
					updateLpa(lpa)

					lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
					lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}
				}

				return appData.Redirect(w, r, lpa, redirect.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}
