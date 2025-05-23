package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

func RemoveTrustCorporation(tmpl template.Template, donorStore DonorStore, isReplacement bool) Handler {
	redirect := donor.PathChooseAttorneysSummary
	titleLabel := "removeTrustCorporation"
	if isReplacement {
		redirect = donor.PathChooseReplacementAttorneysSummary
		titleLabel = "removeReplacementTrustCorporation"
	}

	setTrustCorporation := func(donor *donordata.Provided) {
		donor.Attorneys.TrustCorporation = donordata.TrustCorporation{}
	}

	if isReplacement {
		setTrustCorporation = func(donor *donordata.Provided) {
			donor.ReplacementAttorneys.TrustCorporation = donordata.TrustCorporation{}
		}
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
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
					setTrustCorporation(donor)
					donor.UpdateDecisions()
					donor.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
					donor.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(donor)

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
