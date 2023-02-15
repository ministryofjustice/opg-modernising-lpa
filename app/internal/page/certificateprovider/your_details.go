package certificateprovider

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type cpYourDetailsData struct {
	App        page.AppData
	Lpa        *page.Lpa
	Form       *cpYourDetailsForm
	Errors     validation.List
	DobWarning string
}

type cpYourDetailsForm struct {
	Email            string
	Mobile           string
	Dob              date.Date
	IgnoreDobWarning string
}

func YourDetails(tmpl template.Template, lpaStore page.LpaStore, sessionStore sessions.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProviderSession, err := sesh.CertificateProvider(sessionStore, r)
		if err != nil {
			return err
		}

		if certificateProviderSession.LpaID != lpa.ID {
			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderStart)
		}

		data := &cpYourDetailsData{
			App: appData,
			Lpa: lpa,
			Form: &cpYourDetailsForm{
				Email:  lpa.CertificateProviderProvidedDetails.Email,
				Mobile: lpa.CertificateProviderProvidedDetails.Mobile,
				Dob:    lpa.CertificateProviderProvidedDetails.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readCpYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.None() && data.DobWarning == "" {
				lpa.CertificateProviderProvidedDetails.DateOfBirth = data.Form.Dob
				lpa.CertificateProviderProvidedDetails.Mobile = data.Form.Mobile
				lpa.CertificateProviderProvidedDetails.Email = data.Form.Email

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderYourAddress)
			}
		}

		return tmpl(w, data)
	}
}

func readCpYourDetailsForm(r *http.Request) *cpYourDetailsForm {
	return &cpYourDetailsForm{
		Dob:              date.New(page.PostFormString(r, "date-of-birth-year"), page.PostFormString(r, "date-of-birth-month"), page.PostFormString(r, "date-of-birth-day")),
		Mobile:           page.PostFormString(r, "mobile"),
		Email:            page.PostFormString(r, "email"),
		IgnoreDobWarning: page.PostFormString(r, "ignore-dob-warning"),
	}
}

func (f *cpYourDetailsForm) DobWarning() string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !f.Dob.IsZero() {
		if f.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if f.Dob.Before(today) && f.Dob.After(eighteenYearsEarlier) {
			return "dateOfBirthIsUnder18"
		}
	}

	return ""
}

func (d *cpYourDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.Date("date-of-birth", "dateOfBirth", d.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	errors.String("mobile", "mobile", strings.ReplaceAll(d.Mobile, " ", ""),
		validation.Empty(),
		validation.Mobile())

	errors.String("email", "email", strings.ReplaceAll(d.Email, " ", ""),
		validation.Empty(),
		validation.Email())

	return errors
}
