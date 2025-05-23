package donorpage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

func ChooseReplacementTrustCorporation(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporations, err := reuseStore.TrustCorporations(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}
		if len(trustCorporations) == 0 {
			return donor.PathEnterReplacementTrustCorporation.Redirect(w, r, appData, provided)
		}

		data := &chooseTrustCorporationData{
			App:                 appData,
			Form:                &chooseTrustCorporationForm{},
			Donor:               provided,
			TrustCorporations:   trustCorporations,
			ChooseAttorneysPath: donor.PathEnterReplacementAttorney.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.New {
					return donor.PathEnterReplacementTrustCorporation.Redirect(w, r, appData, provided)
				}

				provided.ReplacementAttorneys.TrustCorporation = trustCorporations[data.Form.Index]
				provided.ReplacementAttorneys.TrustCorporation.UID = newUID()
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := reuseStore.PutTrustCorporation(r.Context(), provided.ReplacementAttorneys.TrustCorporation); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChooseReplacementAttorneysSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
