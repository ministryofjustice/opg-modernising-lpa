package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
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

func EnterDateOfBirth(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		data := &dateOfBirthData{
			App:  appData,
			Lpa:  lpa,
			Form: newDateOfBirthForm(appData.Localizer),
		}

		data.Form.Dob.SetInput(certificateProvider.DateOfBirth)

		if r.Method == http.MethodPost {
			ok := data.Form.Parse(r)
			dobWarning := data.Form.DobWarning()

			if !ok || data.Form.IgnoreDobWarning.Value != dobWarning {
				data.DobWarning = dobWarning
			}

			if ok && data.DobWarning == "" {
				certificateProvider.DateOfBirth = data.Form.Dob.Value
				if !certificateProvider.Tasks.ConfirmYourDetails.IsCompleted() {
					certificateProvider.Tasks.ConfirmYourDetails = task.StateInProgress
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return certificateprovider.PathYourPreferredLanguage.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type dateOfBirthForm struct {
	Dob              *newforms.Date
	IgnoreDobWarning *newforms.String
	Errors           []newforms.Field
}

func newDateOfBirthForm(l Localizer) *dateOfBirthForm {
	return &dateOfBirthForm{
		Dob: newforms.NewDate("date-of-birth", l.T("dateOfBirth")).
			NotEmpty().
			MustBeReal().
			MustBePast().
			BeforeYears(18, l.T("youAreUnder18Error")),
		IgnoreDobWarning: newforms.NewString("ignore-dob-warning", ""),
	}
}

func (f *dateOfBirthForm) Parse(r *http.Request) bool {
	f.Errors = newforms.ParsePostForm(r,
		f.Dob,
		f.IgnoreDobWarning,
	)

	return len(f.Errors) == 0
}

func (f *dateOfBirthForm) DobWarning() string {
	var (
		hundredYearsEarlier = date.Today().AddDate(-100, 0, 0)
	)

	if !f.Dob.Value.IsZero() {
		if f.Dob.Value.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
	}

	return ""
}
