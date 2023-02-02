package page

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
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
			data.Form.Dob = lpa.CertificateProvider.DateOfBirth
		}

		if r.Method == http.MethodPost {
			data.Form = readCertificateProviderDetailsForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.FirstNames = data.Form.FirstNames
				lpa.CertificateProvider.LastName = data.Form.LastName
				lpa.CertificateProvider.DateOfBirth = data.Form.Dob
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
	FirstNames string
	LastName   string
	Dob        date.Date
	Mobile     string
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	d := &certificateProviderDetailsForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Dob = date.New(
		postFormString(r, "date-of-birth-year"),
		postFormString(r, "date-of-birth-month"),
		postFormString(r, "date-of-birth-day"))
	d.Mobile = postFormString(r, "mobile")

	return d
}

func (d *certificateProviderDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", d.FirstNames,
		validation.Empty())

	errors.String("last-name", "lastName", d.LastName,
		validation.Empty())

	errors.Date("date-of-birth", "dateOfBirth", d.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	errors.String("mobile", "mobile", strings.ReplaceAll(d.Mobile, " ", ""),
		validation.Empty(),
		validation.Mobile())

	return errors
}
