package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReplacementTrustCorporationData struct {
	App                            appcontext.Data
	Errors                         validation.List
	Form                           *enterTrustCorporationForm
	LpaID                          string
	ChooseReplacementAttorneysPath string
}

func EnterReplacementTrustCorporation(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporation := provided.ReplacementAttorneys.TrustCorporation

		data := &enterReplacementTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:          trustCorporation.Name,
				CompanyNumber: trustCorporation.CompanyNumber,
				Email:         trustCorporation.Email,
			},
			LpaID:                          provided.LpaID,
			ChooseReplacementAttorneysPath: donor.PathEnterReplacementAttorney.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
				trustCorporation.CompanyNumber = data.Form.CompanyNumber
				trustCorporation.Email = data.Form.Email
				provided.ReplacementAttorneys.TrustCorporation = trustCorporation

				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if trustCorporation.Address.Line1 != "" {
					if err := reuseStore.PutTrustCorporation(r.Context(), trustCorporation); err != nil {
						return err
					}
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathEnterReplacementTrustCorporationAddress.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
