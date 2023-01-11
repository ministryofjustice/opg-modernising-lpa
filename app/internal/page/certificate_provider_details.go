package page

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type certificateProviderDetailsData struct {
	App    AppData
	Errors map[string]string
	Form   *certificateProviderDetailsForm
}

type certificateProviderDetailsForm struct {
	FirstNames       string
	LastName         string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
	Mobile           string
}

func CertificateProviderDetails(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames: lpa.CertificateProvider.FirstNames,
				LastName:   lpa.CertificateProvider.LastName,
				Mobile:     lpa.CertificateProvider.Mobile,
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
				lpa.CertificateProvider.DateOfBirth = data.Form.DateOfBirth
				lpa.CertificateProvider.Mobile = data.Form.Mobile

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole)
			}
		}

		return tmpl(w, data)
	}
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	d := &certificateProviderDetailsForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Dob = Date{
		Day:   postFormString(r, "date-of-birth-day"),
		Month: postFormString(r, "date-of-birth-month"),
		Year:  postFormString(r, "date-of-birth-year"),
	}
	d.Mobile = postFormString(r, "mobile")

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

	isUkMobile, _ := regexp.MatchString(`^(?:07|\+?447)\d{9}$`, strings.ReplaceAll(d.Mobile, " ", ""))

	if !isUkMobile {
		errors["mobile"] = "enterUkMobile"
	}
	if d.Mobile == "" {
		errors["mobile"] = "enterMobile"
	}

	return errors
}
