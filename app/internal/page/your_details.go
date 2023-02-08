package page

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App         AppData
	Errors      validation.List
	Form        *yourDetailsForm
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func YourDetails(tmpl template.Template, lpaStore LpaStore, sessionStore sessions.Store) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourDetailsData{
			App: appData,
			Form: &yourDetailsForm{
				FirstNames: lpa.You.FirstNames,
				LastName:   lpa.You.LastName,
				OtherNames: lpa.You.OtherNames,
				Dob:        lpa.You.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			session, err := sessionStore.Get(r, "session")
			if err != nil {
				return err
			}

			email, ok := session.Values["email"].(string)
			if !ok {
				return fmt.Errorf("no email found in session")
			}

			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeDonor,
				donorMatches(lpa, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if !data.Errors.Any() && data.DobWarning == "" && data.NameWarning == nil {
				lpa.You.FirstNames = data.Form.FirstNames
				lpa.You.LastName = data.Form.LastName
				lpa.You.OtherNames = data.Form.OtherNames
				lpa.You.DateOfBirth = data.Form.Dob
				lpa.You.Email = email
				lpa.Tasks.YourDetails = TaskInProgress

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.YourAddress)
			}
		}

		return tmpl(w, data)
	}
}

type yourDetailsForm struct {
	FirstNames        string
	LastName          string
	OtherNames        string
	Dob               date.Date
	IgnoreDobWarning  string
	IgnoreNameWarning string
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	d := &yourDetailsForm{}

	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.OtherNames = postFormString(r, "other-names")

	d.Dob = date.New(
		postFormString(r, "date-of-birth-year"),
		postFormString(r, "date-of-birth-month"),
		postFormString(r, "date-of-birth-day"))

	d.IgnoreDobWarning = postFormString(r, "ignore-dob-warning")
	d.IgnoreNameWarning = postFormString(r, "ignore-name-warning")

	return d
}

func (f *yourDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("other-names", "otherNamesLabel", f.OtherNames,
		validation.StringTooLong(50))

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	return errors
}

func (f *yourDetailsForm) DobWarning() string {
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
			return "dateOfBirthIsUnder18"
		}
	}

	return ""
}

func donorMatches(lpa *Lpa, firstNames, lastName string) actor.Type {
	for _, attorney := range lpa.Attorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeReplacementAttorney
		}
	}

	if lpa.CertificateProvider.FirstNames == firstNames && lpa.CertificateProvider.LastName == lastName {
		return actor.TypeCertificateProvider
	}

	for _, person := range lpa.PeopleToNotify {
		if person.FirstNames == firstNames && person.LastName == lastName {
			return actor.TypePersonToNotify
		}
	}

	return actor.TypeNone
}
