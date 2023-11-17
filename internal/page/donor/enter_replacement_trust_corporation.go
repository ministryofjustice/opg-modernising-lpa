package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReplacementTrustCorporationData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterTrustCorporationForm
	LpaID  string
}

func EnterReplacementTrustCorporation(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		trustCorporation := lpa.ReplacementAttorneys.TrustCorporation

		data := &enterReplacementTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:          trustCorporation.Name,
				CompanyNumber: trustCorporation.CompanyNumber,
				Email:         trustCorporation.Email,
			},
			LpaID: lpa.ID,
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
				trustCorporation.CompanyNumber = data.Form.CompanyNumber
				trustCorporation.Email = data.Form.Email
				lpa.ReplacementAttorneys.TrustCorporation = trustCorporation

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Paths.EnterReplacementTrustCorporationAddress.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}
