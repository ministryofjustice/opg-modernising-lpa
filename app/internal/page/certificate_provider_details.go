package page

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderDetailsData struct {
	App    AppData
	Errors validation.List
	Form   *certificateProviderDetailsForm
}

func CertificateProviderDetails(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
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

			if data.Errors.None() {
				lpa.CertificateProvider.FirstNames = data.Form.FirstNames
				lpa.CertificateProvider.LastName = data.Form.LastName
				lpa.CertificateProvider.DateOfBirth = data.Form.DateOfBirth
				lpa.CertificateProvider.Mobile = data.Form.Mobile

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole)
			}
		}

		return tmpl(w, data)
	}
}

type certificateProviderDetailsForm struct {
	FirstNames       string
	LastName         string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
	Mobile           string
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

var mobileRegex = regexp.MustCompile(`^(?:07|\+?447)\d{9}$`)

func (d *certificateProviderDetailsForm) Validate() validation.List {
	var errors validation.List

	if d.FirstNames == "" {
		errors.Add("first-names", "enterCertificateProviderFirstNames")
	}
	if d.LastName == "" {
		errors.Add("last-name", "enterCertificateProviderLastName")
	}
	if !d.Dob.Entered() {
		errors.Add("date-of-birth", "enterCertificateProviderDateOfBirth")
	}
	if d.DateOfBirthError != nil {
		errors.Add("date-of-birth", "dateOfBirthMustBeReal")
	}

	if d.Mobile == "" {
		errors.Add("mobile", "enterCertificateProviderMobile")
	}
	if !mobileRegex.MatchString(strings.ReplaceAll(d.Mobile, " ", "")) {
		errors.Add("mobile", "enterUkMobile")
	}

	return errors
}
