package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

func RemoveTrustCorporation(tmpl template.Template, service AttorneyService) Handler {
	redirect := donor.PathChooseAttorneysSummary
	titleLabel := "removeTrustCorporation"
	if service.IsReplacement() {
		redirect = donor.PathChooseReplacementAttorneysSummary
		titleLabel = "removeReplacementTrustCorporation"
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		name := donor.Attorneys.TrustCorporation.Name
		if service.IsReplacement() {
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
					if err := service.DeleteTrustCorporation(r.Context(), donor); err != nil {
						return err
					}
				}

				return redirect.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
