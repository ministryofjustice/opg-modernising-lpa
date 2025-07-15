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

func ChooseTrustCorporation(tmpl template.Template, service AttorneyService, newUID func() actoruid.UID) Handler {
	enterAttorneyPath := donor.PathEnterAttorney
	enterTrustCorporationPath := donor.PathEnterTrustCorporation
	summaryPath := donor.PathChooseAttorneysSummary
	if service.IsReplacement() {
		enterAttorneyPath = donor.PathEnterReplacementAttorney
		enterTrustCorporationPath = donor.PathEnterReplacementTrustCorporation
		summaryPath = donor.PathChooseReplacementAttorneysSummary
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporations, err := service.ReusableTrustCorporations(r.Context(), provided)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}
		if len(trustCorporations) == 0 {
			return enterTrustCorporationPath.Redirect(w, r, appData, provided)
		}

		data := &chooseTrustCorporationData{
			App:                 appData,
			Form:                &chooseTrustCorporationForm{},
			Donor:               provided,
			TrustCorporations:   trustCorporations,
			ChooseAttorneysPath: enterAttorneyPath.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.New {
					return enterTrustCorporationPath.Redirect(w, r, appData, provided)
				}

				if err := service.PutTrustCorporation(r.Context(), provided, trustCorporations[data.Form.Index]); err != nil {
					return err
				}

				return summaryPath.Redirect(w, r, appData, provided)
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
