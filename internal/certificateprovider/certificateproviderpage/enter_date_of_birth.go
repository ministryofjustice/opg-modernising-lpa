package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dateOfBirthData struct {
	App        appcontext.Data
	Lpa        *lpadata.Lpa
	Form       *dateOfBirthForm
	Errors     validation.List
	DobWarning string
}

type dateOfBirthForm struct {
	Dob              date.Date
	IgnoreDobWarning string
}

func EnterDateOfBirth(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &dateOfBirthData{
			App: appData,
			Lpa: lpa,
			Form: &dateOfBirthForm{
				Dob: certificateProvider.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readDateOfBirthForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.None() && data.DobWarning == "" {
				certificateProvider.DateOfBirth = data.Form.Dob
				if !certificateProvider.Tasks.ConfirmYourDetails.IsCompleted() {
					certificateProvider.Tasks.ConfirmYourDetails = task.StateInProgress
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				if lpa.CertificateProvider.Relationship.IsProfessionally() {
					return certificateprovider.PathWhatIsYourHomeAddress.Redirect(w, r, appData, certificateProvider.LpaID)
				}

				return certificateprovider.PathYourPreferredLanguage.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

func readDateOfBirthForm(r *http.Request) *dateOfBirthForm {
	return &dateOfBirthForm{
		Dob:              date.New(page.PostFormString(r, "date-of-birth-year"), page.PostFormString(r, "date-of-birth-month"), page.PostFormString(r, "date-of-birth-day")),
		IgnoreDobWarning: page.PostFormString(r, "ignore-dob-warning"),
	}
}

func (f *dateOfBirthForm) DobWarning() string {
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

func (f *dateOfBirthForm) Validate() validation.List {
	var errors validation.List

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	if f.Dob.After(date.Today().AddDate(-18, 0, 0)) {
		errors.Add("date-of-birth", validation.CustomError{Label: "youAreUnder18Error"})
	}

	return errors
}
