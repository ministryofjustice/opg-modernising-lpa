package page

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App        AppData
	Errors     validation.List
	Form       *yourDetailsForm
	DobWarning string
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
			},
		}

		if !lpa.You.DateOfBirth.IsZero() {
			data.Form.Dob = readDate(lpa.You.DateOfBirth)
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

			if data.Errors.Any() || data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if !data.Errors.Any() && data.DobWarning == "" {
				lpa.You.FirstNames = data.Form.FirstNames
				lpa.You.LastName = data.Form.LastName
				lpa.You.OtherNames = data.Form.OtherNames
				lpa.You.DateOfBirth = data.Form.DateOfBirth
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
	FirstNames       string
	LastName         string
	OtherNames       string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
	IgnoreWarning    string
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	d := &yourDetailsForm{}

	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.OtherNames = postFormString(r, "other-names")

	d.Dob = Date{
		Day:   postFormString(r, "date-of-birth-day"),
		Month: postFormString(r, "date-of-birth-month"),
		Year:  postFormString(r, "date-of-birth-year"),
	}
	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.Dob.Year+"-"+d.Dob.Month+"-"+d.Dob.Day)

	d.IgnoreWarning = postFormString(r, "ignore-warning")

	return d
}

func (d *yourDetailsForm) Validate() validation.List {
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

	if len(d.OtherNames) > 50 {
		errors.Add("other-names", "otherNamesTooLong")
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

func (d *yourDetailsForm) DobWarning() string {
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
			return "dateOfBirthIsUnder18"
		}
	}

	return ""
}
