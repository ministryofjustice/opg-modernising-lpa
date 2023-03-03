package donor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App         page.AppData
	Errors      validation.List
	Form        *yourDetailsForm
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func YourDetails(tmpl template.Template, lpaStore LpaStore, sessionStore sessions.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourDetailsData{
			App: appData,
			Form: &yourDetailsForm{
				FirstNames: lpa.Donor.FirstNames,
				LastName:   lpa.Donor.LastName,
				OtherNames: lpa.Donor.OtherNames,
				Dob:        lpa.Donor.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			donorSession, err := sesh.Donor(sessionStore, r)
			if err != nil {
				return err
			}
			if donorSession.Email == "" {
				return fmt.Errorf("no email in session session")
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
				lpa.Donor.FirstNames = data.Form.FirstNames
				lpa.Donor.LastName = data.Form.LastName
				lpa.Donor.OtherNames = data.Form.OtherNames
				lpa.Donor.DateOfBirth = data.Form.Dob
				lpa.Donor.Email = donorSession.Email
				lpa.Tasks.YourDetails = page.TaskInProgress

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.YourAddress)
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

	d.FirstNames = page.PostFormString(r, "first-names")
	d.LastName = page.PostFormString(r, "last-name")
	d.OtherNames = page.PostFormString(r, "other-names")

	d.Dob = date.New(
		page.PostFormString(r, "date-of-birth-year"),
		page.PostFormString(r, "date-of-birth-month"),
		page.PostFormString(r, "date-of-birth-day"))

	d.IgnoreDobWarning = page.PostFormString(r, "ignore-dob-warning")
	d.IgnoreNameWarning = page.PostFormString(r, "ignore-name-warning")

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

func donorMatches(lpa *page.Lpa, firstNames, lastName string) actor.Type {
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
