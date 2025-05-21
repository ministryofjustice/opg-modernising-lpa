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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Form                     *chooseAttorneysForm
	Donor                    *donordata.Provided
	Attorneys                []donordata.Attorney
	ShowTrustCorporationLink bool
}

func ChooseReplacementAttorneys(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorneys, err := reuseStore.Attorneys(r.Context(), provided)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}
		if len(attorneys) == 0 {
			return donor.PathEnterReplacementAttorney.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
		}

		data := &chooseAttorneysData{
			App:                      appData,
			Form:                     &chooseAttorneysForm{},
			Donor:                    provided,
			Attorneys:                attorneys,
			ShowTrustCorporationLink: provided.CanAddTrustCorporation(),
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)

			if data.Errors.None() {
				if len(data.Form.Indices) == 0 {
					return donor.PathEnterReplacementAttorney.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
				}

				for _, index := range data.Form.Indices {
					attorney := attorneys[index]
					attorney.UID = newUID()
					provided.ReplacementAttorneys.Attorneys = append(provided.ReplacementAttorneys.Attorneys, attorney)
				}

				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := reuseStore.PutAttorneys(r.Context(), provided.ReplacementAttorneys.Attorneys); err != nil {
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
