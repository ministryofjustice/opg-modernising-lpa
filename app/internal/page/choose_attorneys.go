package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysData struct {
	App         AppData
	Errors      validation.List
	Form        *chooseAttorneysForm
	ShowDetails bool
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func ChooseAttorneys(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := lpa.GetAttorney(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.Attorneys) > 0 && attorneyFound == false && addAnother == false {
			return appData.Redirect(w, r, lpa, Paths.ChooseAttorneysSummary)
		}

		data := &chooseAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
			ShowDetails: attorneyFound == false && addAnother == false,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeAttorney,
				attorneyMatches(lpa, attorney.ID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.Dob,
						ID:          randomString(8),
					}

					lpa.Attorneys = append(lpa.Attorneys, attorney)
				} else {
					attorney.FirstNames = data.Form.FirstNames
					attorney.LastName = data.Form.LastName
					attorney.Email = data.Form.Email
					attorney.DateOfBirth = data.Form.Dob

					lpa.PutAttorney(attorney)
				}

				if !attorneyFound {
					lpa.Tasks.ChooseAttorneys = TaskInProgress
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")
				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChooseAttorneysAddress, attorney.ID)
				}

				return appData.Redirect(w, r, lpa, from)
			}
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
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")
	d.Dob = date.New(
		postFormString(r, "date-of-birth-year"),
		postFormString(r, "date-of-birth-month"),
		postFormString(r, "date-of-birth-day"))

	d.IgnoreDobWarning = postFormString(r, "ignore-dob-warning")
	d.IgnoreNameWarning = postFormString(r, "ignore-name-warning")

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
		validation.Empty(),
		validation.Email())

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	return errors
}

func (d *chooseAttorneysForm) DobWarning() string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !d.Dob.IsZero() {
		if d.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if d.Dob.Before(today) && d.Dob.After(eighteenYearsEarlier) {
			return "attorneyDateOfBirthIsUnder18"
		}
	}

	return ""
}

func attorneyMatches(lpa *Lpa, id, firstNames, lastName string) actor.Type {
	if lpa.You.FirstNames == firstNames && lpa.You.LastName == lastName {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys {
		if attorney.ID != id && attorney.FirstNames == firstNames && attorney.LastName == lastName {
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
