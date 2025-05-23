package donorpage

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseTrustCorporationData struct {
	App                 appcontext.Data
	Errors              validation.List
	Form                *chooseTrustCorporationForm
	Donor               *donordata.Provided
	TrustCorporations   []donordata.TrustCorporation
	ChooseAttorneysPath string
}

func ChooseTrustCorporation(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporations, err := reuseStore.TrustCorporations(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}
		if len(trustCorporations) == 0 {
			return donor.PathEnterTrustCorporation.Redirect(w, r, appData, provided)
		}

		data := &chooseTrustCorporationData{
			App:                 appData,
			Form:                &chooseTrustCorporationForm{},
			Donor:               provided,
			TrustCorporations:   trustCorporations,
			ChooseAttorneysPath: donor.PathEnterAttorney.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.New {
					return donor.PathEnterTrustCorporation.Redirect(w, r, appData, provided)
				}

				provided.Attorneys.TrustCorporation = trustCorporations[data.Form.Index]
				provided.Attorneys.TrustCorporation.UID = newUID()

				provided.UpdateDecisions()
				provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := reuseStore.PutTrustCorporation(r.Context(), provided.Attorneys.TrustCorporation); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChooseAttorneysSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type chooseTrustCorporationForm struct {
	New   bool
	Index int
	Err   error
}

func readChooseTrustCorporationForm(r *http.Request) *chooseTrustCorporationForm {
	option := page.PostFormString(r, "option")
	index, err := strconv.Atoi(option)

	return &chooseTrustCorporationForm{
		New:   option == "new",
		Index: index,
		Err:   err,
	}
}

func (f *chooseTrustCorporationForm) Validate() validation.List {
	var errors validation.List

	if !f.New && f.Err != nil {
		errors.Add("option", validation.SelectError{Label: "aTrustCorporationOrToAddANewTrustCorporation"})
	}

	return errors
}
