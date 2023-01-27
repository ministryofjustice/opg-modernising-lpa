package page

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysData struct {
	App         AppData
	Errors      validation.List
	Form        *chooseAttorneysForm
	ShowDetails bool
	DobWarning  string
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
			},
			ShowDetails: attorneyFound == false && addAnother == false,
		}

		if !attorney.DateOfBirth.IsZero() {
			data.Form.Dob = readDate(attorney.DateOfBirth)
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Empty() && data.DobWarning == "" {
				if attorneyFound == false {
					attorney = Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.DateOfBirth,
						ID:          randomString(8),
					}

					lpa.Attorneys = append(lpa.Attorneys, attorney)
				} else {
					attorney.FirstNames = data.Form.FirstNames
					attorney.LastName = data.Form.LastName
					attorney.Email = data.Form.Email
					attorney.DateOfBirth = data.Form.DateOfBirth

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
	FirstNames       string
	LastName         string
	Email            string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
	IgnoreWarning    string
}

func readChooseAttorneysForm(r *http.Request) *chooseAttorneysForm {
	d := &chooseAttorneysForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")
	d.Dob = Date{
		Day:   postFormString(r, "date-of-birth-day"),
		Month: postFormString(r, "date-of-birth-month"),
		Year:  postFormString(r, "date-of-birth-year"),
	}

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.Dob.Year+"-"+d.Dob.Month+"-"+d.Dob.Day)

	d.IgnoreWarning = postFormString(r, "ignore-warning")

	return d
}

func (d *chooseAttorneysForm) Validate() validation.List {
	var errors validation.List

	if d.FirstNames == "" {
		errors.Add("first-names", "enterFirstNames")
	}
	if len(d.FirstNames) > 53 {
		errors.Add("first-names", "firstNamesTooLong")
	}

	if d.LastName == "" {
		errors.Add("last-name", "enterLastName")
	}
	if len(d.LastName) > 61 {
		errors.Add("last-name", "lastNameTooLong")
	}

	if d.Email == "" {
		errors.Add("email", "enterEmail")
	} else if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", d.Email)); err != nil {
		errors.Add("email", "emailIncorrectFormat")
	}

	if d.Dob.Day == "" || d.Dob.Month == "" || d.Dob.Year == "" {
		errors.Add("date-of-birth", "enterDateOfBirth")
	} else if d.DateOfBirthError != nil {
		errors.Add("date-of-birth", "dateOfBirthMustBeReal")
	} else {
		today := time.Now().UTC().Round(24 * time.Hour)

		if d.DateOfBirth.After(today) {
			errors.Add("date-of-birth", "dateOfBirthIsFuture")
		}
	}

	return errors
}

func (d *chooseAttorneysForm) DobWarning() string {
	var (
		today                = time.Now().UTC().Round(24 * time.Hour)
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !d.DateOfBirth.IsZero() {
		if d.DateOfBirth.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if d.DateOfBirth.Before(today) && d.DateOfBirth.After(eighteenYearsEarlier) {
			return "attorneyDateOfBirthIsUnder18"
		}
	}

	return ""
}
