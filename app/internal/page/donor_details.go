package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type donorDetailsData struct {
	Page   string
	L      localize.Localizer
	Lang   Lang
	Errors map[string]string
	Form   *donorDetailsForm
}

func DonorDetails(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, dataStore DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &donorDetailsData{
			Page: "/donor-details",
			L:    localizer,
			Lang: lang,
			Form: &donorDetailsForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readDonorDetailsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				dataStore.Save(Donor{
					FirstName:   data.Form.FirstName,
					LastName:    data.Form.LastName,
					DateOfBirth: data.Form.DateOfBirth,
				})
				lang.Redirect(w, r, "/donor-address", http.StatusFound)
				return
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}

type donorDetailsForm struct {
	FirstName        string
	LastName         string
	DobDay           string
	DobMonth         string
	DobYear          string
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readDonorDetailsForm(r *http.Request) *donorDetailsForm {
	d := &donorDetailsForm{}
	d.FirstName = postFormString(r, "first-name")
	d.LastName = postFormString(r, "last-name")
	d.DobDay = postFormString(r, "date-of-birth-day")
	d.DobMonth = postFormString(r, "date-of-birth-month")
	d.DobYear = postFormString(r, "date-of-birth-year")

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.DobYear+"-"+d.DobMonth+"-"+d.DobDay)

	return d
}

func (d *donorDetailsForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstName == "" {
		errors["first-name"] = "enterYourFirstName"
	}
	if d.LastName == "" {
		errors["last-name"] = "enterYourLastName"
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
		errors["date-of-birth"] = "yourDateOfBirthMustBeReal"
	}

	return errors
}
