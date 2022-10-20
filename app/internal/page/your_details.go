package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type yourDetailsData struct {
	App        AppData
	Errors     map[string]string
	Form       *yourDetailsForm
	DobWarning string
}

func YourDetails(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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
			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if len(data.Errors) == 0 && data.DobWarning == "" {
				lpa.You.FirstNames = data.Form.FirstNames
				lpa.You.LastName = data.Form.LastName
				lpa.You.OtherNames = data.Form.OtherNames
				lpa.You.DateOfBirth = data.Form.DateOfBirth

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, yourAddressPath, http.StatusFound)
				return nil
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

func (d *yourDetailsForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstNames == "" {
		errors["first-names"] = "enterFirstNames"
	}
	if len(d.FirstNames) > 53 {
		errors["first-names"] = "firstNamesTooLong"
	}

	if d.LastName == "" {
		errors["last-name"] = "enterLastName"
	}
	if len(d.LastName) > 61 {
		errors["last-name"] = "lastNameTooLong"
	}

	if len(d.OtherNames) > 50 {
		errors["other-names"] = "otherNamesTooLong"
	}

	if d.Dob.Day == "" || d.Dob.Month == "" || d.Dob.Year == "" {
		errors["date-of-birth"] = "enterDateOfBirth"
	} else if d.DateOfBirthError != nil {
		errors["date-of-birth"] = "dateOfBirthMustBeReal"
	} else {
		today := time.Now().UTC().Round(24 * time.Hour)

		if d.DateOfBirth.After(today) {
			errors["date-of-birth"] = "dateOfBirthIsFuture"
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
