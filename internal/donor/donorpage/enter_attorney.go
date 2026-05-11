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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterAttorneyData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Donor                    *donordata.Provided
	Form                     *enterAttorneyForm
	ShowTrustCorporationLink bool
}

func EnterAttorney(tmpl template.Template, service AttorneyService) Handler {
	matches := attorneyMatches
	addressPath := donor.PathChooseAttorneysAddress
	actorType := actor.TypeAttorney
	if service.IsReplacement() {
		matches = replacementAttorneyMatches
		addressPath = donor.PathChooseReplacementAttorneysAddress
		actorType = actor.TypeReplacementAttorney
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		uid := actoruid.FromRequest(r)

		if uid.IsZero() {
			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		var (
			attorney      donordata.Attorney
			attorneyFound bool
		)
		if service.IsReplacement() {
			attorney, attorneyFound = provided.ReplacementAttorneys.Get(uid)
		} else {
			attorney, attorneyFound = provided.Attorneys.Get(uid)
		}

		data := &enterAttorneyData{
			App:   appData,
			Donor: provided,
			Form: &enterAttorneyForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
			ShowTrustCorporationLink: service.CanAddTrustCorporation(provided),
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterAttorneyForm(r)
			data.Errors = data.Form.Validate()
			redirectToWarning := false

			nameMatches := matches(provided, attorney.UID, data.Form.FirstNames, data.Form.LastName)
			if attorney.NameHasChanged(data.Form.FirstNames, data.Form.LastName) && !nameMatches.IsNone() {
				redirectToWarning = true
			}

			dobWarning := dateOfBirthWarning(data.Form.Dob.Value, actorType)
			if (data.Form.Dob.Value != attorney.DateOfBirth || attorney.DateOfBirth.After(date.Today().AddDate(-18, 0, 0))) && dobWarning != "" {
				redirectToWarning = true
			}

			if data.Errors.None() {
				if attorneyFound == false {
					attorney = donordata.Attorney{UID: uid}
				}

				attorney.FirstNames = data.Form.FirstNames.Value
				attorney.LastName = data.Form.LastName.Value
				attorney.Email = data.Form.Email.Value
				attorney.DateOfBirth = data.Form.Dob.Value

				if err := service.Put(r.Context(), provided, attorney); err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"id":          {attorney.UID.String()},
						"warningFrom": {appData.Page},
						"next": {addressPath.FormatQuery(
							provided.LpaID,
							url.Values{"id": {attorney.UID.String()}}),
						},
						"actor": {actorType.String()},
					})
				}

				return addressPath.RedirectQuery(w, r, appData, provided, url.Values{"id": {attorney.UID.String()}})
			}
		}

		return tmpl(w, data)
	}
}

type enterAttorneyForm struct {
	newforms.Form
	FirstNames *newforms.String
	LastName   *newforms.String
	Email      *newforms.String
	Dob        *newforms.Date
}

func newEnterAttorneyForm(l Localizer) *enterAttorneyForm {
	return &enterAttorneyForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
		Email: newforms.NewString("email", l.T("email")).
			Email(),
		Dob: newforms.NewDate("date-of-birth", l.T("dateOfBirth")).
			NotEmpty().
			MustBePast().
			MustBeReal(),
	}
}

func (f *enterAttorneyForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
		f.Email,
		f.Dob,
	)
}
