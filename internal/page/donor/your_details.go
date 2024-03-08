package donor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App               page.AppData
	Errors            validation.List
	Form              *yourDetailsForm
	YesNoMaybeOptions actor.YesNoMaybeOptions
	DobWarning        string
	NameWarning       *actor.SameNameWarning
}

func YourDetails(tmpl template.Template, donorStore DonorStore, sessionStore SessionStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourDetailsData{
			App: appData,
			Form: &yourDetailsForm{
				FirstNames: donor.Donor.FirstNames,
				LastName:   donor.Donor.LastName,
				OtherNames: donor.Donor.OtherNames,
				Dob:        donor.Donor.DateOfBirth,
				CanSign:    donor.Donor.ThinksCanSign,
			},
			YesNoMaybeOptions: actor.YesNoMaybeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeDonor,
				donorMatches(donor, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() ||
				data.Form.IgnoreNameWarning != nameWarning.String() &&
					donor.Donor.FullName() != fmt.Sprintf("%s %s", data.Form.FirstNames, data.Form.LastName) {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				redirect := page.Paths.YourAddress

				if donor.Donor.FirstNames != data.Form.FirstNames || donor.Donor.LastName != data.Form.LastName || donor.Donor.DateOfBirth != data.Form.Dob {
					donor.Donor.FirstNames = data.Form.FirstNames
					donor.Donor.LastName = data.Form.LastName
					donor.Donor.DateOfBirth = data.Form.Dob

					donor.HasSentApplicationUpdatedEvent = false
				}

				donor.Donor.OtherNames = data.Form.OtherNames
				donor.Donor.ThinksCanSign = data.Form.CanSign

				if appData.IsSupporter {
					donor.Donor.Email = data.Form.Email
				} else {
					loginSession, err := sessionStore.Login(r)
					if err != nil {
						return err
					}
					if loginSession.Email == "" {
						return fmt.Errorf("no email in login session")
					}

					donor.Donor.Email = loginSession.Email
				}

				if donor.Donor.ThinksCanSign.IsYes() {
					donor.Donor.CanSign = form.Yes
				} else {
					redirect = page.Paths.CheckYouCanSign
				}

				if !donor.Tasks.YourDetails.Completed() {
					donor.Tasks.YourDetails = actor.TaskInProgress
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, donor)
			}
		}

		if !donor.Donor.DateOfBirth.IsZero() {
			data.DobWarning = data.Form.DobWarning()
		}

		return tmpl(w, data)
	}
}

type yourDetailsForm struct {
	FirstNames        string
	LastName          string
	OtherNames        string
	Email             string
	Dob               date.Date
	CanSign           actor.YesNoMaybe
	CanSignError      error
	IgnoreDobWarning  string
	IgnoreNameWarning string
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	d := &yourDetailsForm{}

	d.FirstNames = page.PostFormString(r, "first-names")
	d.LastName = page.PostFormString(r, "last-name")
	d.OtherNames = page.PostFormString(r, "other-names")
	d.Email = page.PostFormString(r, "email")

	d.Dob = date.New(
		page.PostFormString(r, "date-of-birth-year"),
		page.PostFormString(r, "date-of-birth-month"),
		page.PostFormString(r, "date-of-birth-day"))

	d.CanSign, d.CanSignError = actor.ParseYesNoMaybe(page.PostFormString(r, "can-sign"))

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

	errors.String("email", "email", f.Email,
		validation.Email())

	errors.String("other-names", "otherNamesYouAreKnownBy", f.OtherNames,
		validation.StringTooLong(50))

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	errors.Error("can-sign", "yesIfCanSign", f.CanSignError,
		validation.Selected())

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

func donorMatches(donor *actor.DonorProvidedDetails, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(donor.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(donor.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	for _, person := range donor.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, firstNames) && strings.EqualFold(person.LastName, lastName) {
			return actor.TypePersonToNotify
		}
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, firstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, lastName) {
		return actor.TypeAuthorisedSignatory
	}

	if strings.EqualFold(donor.IndependentWitness.FirstNames, firstNames) && strings.EqualFold(donor.IndependentWitness.LastName, lastName) {
		return actor.TypeIndependentWitness
	}

	return actor.TypeNone
}
