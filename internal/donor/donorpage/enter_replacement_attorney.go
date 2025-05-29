package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReplacementAttorneyData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Donor                    *donordata.Provided
	Form                     *enterAttorneyForm
	DobWarning               string
	NameWarning              *actor.SameNameWarning
	ShowTrustCorporationLink bool
}

func EnterReplacementAttorney(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		uid := actoruid.FromRequest(r)

		if uid.IsZero() {
			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		attorney, attorneyFound := provided.ReplacementAttorneys.Get(uid)

		data := &enterReplacementAttorneyData{
			App:   appData,
			Donor: provided,
			Form: &enterAttorneyForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
			ShowTrustCorporationLink: provided.CanAddTrustCorporation(),
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterAttorneyForm(r)
			data.Errors = data.Form.Validate()
			redirectToWarning := false

			nameMatches := replacementAttorneyMatches(provided, attorney.UID, data.Form.FirstNames, data.Form.LastName)

			if attorney.NameHasChanged(data.Form.FirstNames, data.Form.LastName) && !nameMatches.IsNone() {
				redirectToWarning = true
			}

			dobWarning := dateOfBirthWarning(data.Form.Dob, actor.TypeReplacementAttorney)

			if (data.Form.Dob != attorney.DateOfBirth || attorney.DateOfBirth.After(date.Today().AddDate(-18, 0, 0))) && dobWarning != "" {
				redirectToWarning = true
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = donordata.Attorney{UID: uid}
				}

				attorney.FirstNames = data.Form.FirstNames
				attorney.LastName = data.Form.LastName
				attorney.Email = data.Form.Email
				attorney.DateOfBirth = data.Form.Dob

				provided.ReplacementAttorneys.Put(attorney)

				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := reuseStore.PutAttorney(r.Context(), attorney); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"id":          {attorney.UID.String()},
						"warningFrom": {appData.Page},
						"next": {donor.PathChooseReplacementAttorneysAddress.FormatQuery(
							provided.LpaID,
							url.Values{"id": {attorney.UID.String()}}),
						},
						"actor": {actor.TypeReplacementAttorney.String()},
					})
				}

				return donor.PathChooseReplacementAttorneysAddress.RedirectQuery(w, r, appData, provided, url.Values{"id": {attorney.UID.String()}})
			}
		}

		return tmpl(w, data)
	}
}
