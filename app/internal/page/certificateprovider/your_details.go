package certificateprovider

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App        page.AppData
	Lpa        *page.Lpa
	Form       *yourDetailsForm
	Errors     validation.List
	DobWarning string
}

type yourDetailsForm struct {
	Mobile           string
	Dob              date.Date
	IgnoreDobWarning string
}

func YourDetails(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourDetailsData{
			App: appData,
			Lpa: lpa,
			Form: &yourDetailsForm{
				Mobile: lpa.CertificateProviderProvidedDetails.Mobile,
				Dob:    lpa.CertificateProviderProvidedDetails.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.None() && data.DobWarning == "" {
				lpa.CertificateProviderProvidedDetails.DateOfBirth = data.Form.Dob
				lpa.CertificateProviderProvidedDetails.Mobile = data.Form.Mobile

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderYourAddress)
			}
		}

		return tmpl(w, data)
	}
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	return &yourDetailsForm{
		Dob:              date.New(page.PostFormString(r, "date-of-birth-year"), page.PostFormString(r, "date-of-birth-month"), page.PostFormString(r, "date-of-birth-day")),
		Mobile:           page.PostFormString(r, "mobile"),
		IgnoreDobWarning: page.PostFormString(r, "ignore-dob-warning"),
	}
}

func (f *yourDetailsForm) DobWarning() string {
	var (
		hundredYearsEarlier = date.Today().AddDate(-100, 0, 0)
	)

	if !f.Dob.IsZero() {
		if f.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
	}

	return ""
}

func (f *yourDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.Date("date-of-birth", "yourDateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBePast())

	if !f.Dob.Valid() {
		errors.Add("date-of-birth", validation.EnterError{Label: "aValidDateOfBirth"})
	}

	if f.Dob.After(date.Today().AddDate(-18, 0, 0)) {
		errors.Add("date-of-birth", validation.CustomError{Label: "youAreUnder18Error"})
	}

	errors.String("mobile", "yourUkMobile", strings.ReplaceAll(f.Mobile, " ", ""),
		validation.Empty())

	if !validation.MobileRegex.MatchString(f.Mobile) {
		errors.Add("mobile", validation.EnterError{Label: "aValidUkMobileLike"})
	}

	return errors
}
