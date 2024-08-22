package donorpage

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Donor                    *donordata.Provided
	Form                     *chooseAttorneysForm
	ShowDetails              bool
	DobWarning               string
	NameWarning              *actor.SameNameWarning
	ShowTrustCorporationLink bool
}

func ChooseAttorneys(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		uid := actoruid.FromRequest(r)

		if uid.IsZero() {
			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := provided.Attorneys.Get(uid)

		data := &chooseAttorneysData{
			App:   appData,
			Donor: provided,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
			ShowDetails:              attorneyFound == false && addAnother == false,
			ShowTrustCorporationLink: provided.Type.IsPropertyAndAffairs() && provided.ReplacementAttorneys.TrustCorporation.Name == "",
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeAttorney,
				attorneyMatches(provided, attorney.UID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Form.Dob != attorney.DateOfBirth && (data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning) {
				data.DobWarning = dobWarning
			}

			if data.Form.NameHasChanged(attorney) && (data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String()) {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = donordata.Attorney{UID: uid}
				}

				attorney.FirstNames = data.Form.FirstNames
				attorney.LastName = data.Form.LastName
				attorney.Email = data.Form.Email
				attorney.DateOfBirth = data.Form.Dob

				provided.Attorneys.Put(attorney)

				provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChooseAttorneysAddress.RedirectQuery(w, r, appData, provided, url.Values{"id": {attorney.UID.String()}})
			}
		}

		if !attorney.DateOfBirth.IsZero() {
			data.DobWarning = data.Form.DobWarning()
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysForm struct {
	FirstNames        string
	LastName          string
	Email             string
	Dob               date.Date
	IgnoreDobWarning  string
	IgnoreNameWarning string
}

func readChooseAttorneysForm(r *http.Request) *chooseAttorneysForm {
	d := &chooseAttorneysForm{}
	d.FirstNames = page.PostFormString(r, "first-names")
	d.LastName = page.PostFormString(r, "last-name")
	d.Email = page.PostFormString(r, "email")
	d.Dob = date.New(
		page.PostFormString(r, "date-of-birth-year"),
		page.PostFormString(r, "date-of-birth-month"),
		page.PostFormString(r, "date-of-birth-day"))

	d.IgnoreDobWarning = page.PostFormString(r, "ignore-dob-warning")
	d.IgnoreNameWarning = page.PostFormString(r, "ignore-name-warning")

	return d
}

func (f *chooseAttorneysForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("email", "email", f.Email,
		validation.Email())

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	return errors
}

func (f *chooseAttorneysForm) DobWarning() string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !f.Dob.IsZero() {
		if f.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if f.Dob.Before(today) && f.Dob.After(eighteenYearsEarlier) {
			return "attorneyDateOfBirthIsUnder18"
		}
	}

	return ""
}

func attorneyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsAttorney() && person.UID == uid) &&
			strings.EqualFold(person.FirstNames, firstNames) &&
			strings.EqualFold(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func (f *chooseAttorneysForm) NameHasChanged(attorney donordata.Attorney) bool {
	return attorney.FirstNames != f.FirstNames || attorney.LastName != f.LastName
}
