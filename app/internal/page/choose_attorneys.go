package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseAttorneysData struct {
	App    AppData
	Errors map[string]string
	Form   *chooseAttorneysForm
}

func ChooseAttorneys(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &chooseAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: lpa.Attorney.FirstNames,
				LastName:   lpa.Attorney.LastName,
				Email:      lpa.Attorney.Email,
			},
		}

		if !lpa.Attorney.DateOfBirth.IsZero() {
			data.Form.DobDay = lpa.Attorney.DateOfBirth.Format("2")
			data.Form.DobMonth = lpa.Attorney.DateOfBirth.Format("1")
			data.Form.DobYear = lpa.Attorney.DateOfBirth.Format("2006")
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.Attorney.FirstNames = data.Form.FirstNames
				lpa.Attorney.LastName = data.Form.LastName
				lpa.Attorney.Email = data.Form.Email
				lpa.Attorney.DateOfBirth = data.Form.DateOfBirth

				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, chooseAttorneysAddressPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysForm struct {
	FirstNames       string
	LastName         string
	Email            string
	DobDay           string
	DobMonth         string
	DobYear          string
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readChooseAttorneysForm(r *http.Request) *chooseAttorneysForm {
	d := &chooseAttorneysForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")
	d.DobDay = postFormString(r, "date-of-birth-day")
	d.DobMonth = postFormString(r, "date-of-birth-month")
	d.DobYear = postFormString(r, "date-of-birth-year")

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.DobYear+"-"+d.DobMonth+"-"+d.DobDay)

	return d
}

func (d *chooseAttorneysForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstNames == "" {
		errors["first-names"] = "enterFirstNames"
	}
	if d.LastName == "" {
		errors["last-name"] = "enterLastName"
	}
	if d.Email == "" {
		errors["email"] = "enterEmail"
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
