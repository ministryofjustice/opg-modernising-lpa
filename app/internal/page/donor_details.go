package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type donorDetailsData struct {
	App    AppData
	Errors map[string]string
	Form   *donorDetailsForm
}

func DonorDetails(logger Logger, tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) {
		var lpa Lpa
		dataStore.Get(r.Context(), appData.SessionID, &lpa)

		data := &donorDetailsData{
			App: appData,
			Form: &donorDetailsForm{
				FirstNames: lpa.Donor.FirstNames,
				LastName:   lpa.Donor.LastName,
				OtherNames: lpa.Donor.OtherNames,
			},
		}

		if !lpa.Donor.DateOfBirth.IsZero() {
			data.Form.DobDay = lpa.Donor.DateOfBirth.Format("2")
			data.Form.DobMonth = lpa.Donor.DateOfBirth.Format("1")
			data.Form.DobYear = lpa.Donor.DateOfBirth.Format("2006")
		}

		if r.Method == http.MethodPost {
			data.Form = readDonorDetailsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.Donor.FirstNames = data.Form.FirstNames
				lpa.Donor.LastName = data.Form.LastName
				lpa.Donor.OtherNames = data.Form.OtherNames
				lpa.Donor.DateOfBirth = data.Form.DateOfBirth

				dataStore.Put(r.Context(), appData.SessionID, lpa)
				appData.Lang.Redirect(w, r, donorAddressPath, http.StatusFound)
				return
			}
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}

type donorDetailsForm struct {
	FirstNames       string
	LastName         string
	OtherNames       string
	DobDay           string
	DobMonth         string
	DobYear          string
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readDonorDetailsForm(r *http.Request) *donorDetailsForm {
	d := &donorDetailsForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.OtherNames = postFormString(r, "other-names")
	d.DobDay = postFormString(r, "date-of-birth-day")
	d.DobMonth = postFormString(r, "date-of-birth-month")
	d.DobYear = postFormString(r, "date-of-birth-year")

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.DobYear+"-"+d.DobMonth+"-"+d.DobDay)

	return d
}

func (d *donorDetailsForm) Validate() map[string]string {
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
