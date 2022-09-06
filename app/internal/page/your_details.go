package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type yourDetailsData struct {
	App    AppData
	Errors map[string]string
	Form   *yourDetailsForm
}

func YourDetails(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
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
			data.Form.DobDay = lpa.You.DateOfBirth.Format("2")
			data.Form.DobMonth = lpa.You.DateOfBirth.Format("1")
			data.Form.DobYear = lpa.You.DateOfBirth.Format("2006")
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.You.FirstNames = data.Form.FirstNames
				lpa.You.LastName = data.Form.LastName
				lpa.You.OtherNames = data.Form.OtherNames
				lpa.You.DateOfBirth = data.Form.DateOfBirth

				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
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
	DobDay           string
	DobMonth         string
	DobYear          string
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	d := &yourDetailsForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.OtherNames = postFormString(r, "other-names")
	d.DobDay = postFormString(r, "date-of-birth-day")
	d.DobMonth = postFormString(r, "date-of-birth-month")
	d.DobYear = postFormString(r, "date-of-birth-year")

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.DobYear+"-"+d.DobMonth+"-"+d.DobDay)

	return d
}

func (d *yourDetailsForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstNames == "" {
		errors["first-names"] = "enterFirstNames"
	}
	if d.LastName == "" {
		errors["last-name"] = "enterLastName"
	}
	if d.DobDay == "" {
		errors["date-of-birth"] = "dateOfBirthDay"
	}
	if d.DobMonth == "" {
		errors["date-of-birth"] = "dateOfBirthMonth"
	}
	if d.DobYear == "" {
		errors["date-of-birth"] = "dateOfBirthYear"
	}
	if _, ok := errors["date-of-birth"]; !ok && d.DateOfBirthError != nil {
		errors["date-of-birth"] = "dateOfBirthMustBeReal"
	}

	return errors
}
