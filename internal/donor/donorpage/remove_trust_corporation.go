package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func RemoveTrustCorporation(tmpl template.Template, donorStore DonorStore, isReplacement bool) Handler {
	redirect := page.Paths.ChooseAttorneysSummary
	titleLabel := "removeTrustCorporation"
	if isReplacement {
		redirect = page.Paths.ChooseReplacementAttorneysSummary
		titleLabel = "removeReplacementTrustCorporation"
	}

	updateDonor := func(donor *donordata.Provided) {
		donor.Attorneys.TrustCorporation = donordata.TrustCorporation{}
		if donor.Attorneys.Len() == 1 {
			donor.AttorneyDecisions = donordata.AttorneyDecisions{}
		}
	}

	if isReplacement {
		updateDonor = func(donor *donordata.Provided) {
			donor.ReplacementAttorneys.TrustCorporation = donordata.TrustCorporation{}
			if donor.ReplacementAttorneys.Len() == 1 {
				donor.ReplacementAttorneyDecisions = donordata.AttorneyDecisions{}
			}
		}
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		name := donor.Attorneys.TrustCorporation.Name
		if isReplacement {
			name = donor.ReplacementAttorneys.TrustCorporation.Name
		}

		data := &removeAttorneyData{
			App:        appData,
			TitleLabel: titleLabel,
			Name:       name,
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveTrustCorporation")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo.IsYes() {
					updateDonor(donor)

					donor.Tasks.ChooseAttorneys = page.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
					donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				return redirect.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
