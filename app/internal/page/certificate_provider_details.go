package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type certificateProviderDetailsData struct {
	App    AppData
	Errors map[string]string
	Form   *certificateProviderDetailsForm
}

func CertificateProviderDetails(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames: lpa.CertificateProvider.FirstNames,
				LastName:   lpa.CertificateProvider.LastName,
				Email:      lpa.CertificateProvider.Email,
			},
		}

		if !lpa.CertificateProvider.DateOfBirth.IsZero() {
			data.Form.Dob = readDate(lpa.CertificateProvider.DateOfBirth)
		}

		if r.Method == http.MethodPost {
			data.Form = readCertificateProviderDetailsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.CertificateProvider.FirstNames = data.Form.FirstNames
				lpa.CertificateProvider.LastName = data.Form.LastName
				lpa.CertificateProvider.Email = data.Form.Email
				lpa.CertificateProvider.DateOfBirth = data.Form.DateOfBirth

				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, howDoYouKnowYourCertificateProviderPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type certificateProviderDetailsForm struct {
	FirstNames       string
	LastName         string
	Email            string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	d := &certificateProviderDetailsForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")
	d.Dob = Date{
		Day:   postFormString(r, "date-of-birth-day"),
		Month: postFormString(r, "date-of-birth-month"),
		Year:  postFormString(r, "date-of-birth-year"),
	}

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.Dob.Year+"-"+d.Dob.Month+"-"+d.Dob.Day)

	return d
}

func (d *certificateProviderDetailsForm) Validate() map[string]string {
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
	if d.Dob.Day == "" {
		errors["date-of-birth"] = "dateOfBirthDay"
	}
	if d.Dob.Month == "" {
		errors["date-of-birth"] = "dateOfBirthMonth"
	}
	if d.Dob.Year == "" {
		errors["date-of-birth"] = "dateOfBirthYear"
	}
	if _, ok := errors["date-of-birth"]; !ok && d.DateOfBirthError != nil {
		errors["date-of-birth"] = "dateOfBirthMustBeReal"
	}

	return errors
}
