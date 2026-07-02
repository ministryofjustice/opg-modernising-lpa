package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type dateOfBirthData struct {
	App        appcontext.Data
	Lpa        *lpadata.Lpa
	Form       *dateOfBirthForm
	DobWarning string
}

func EnterDateOfBirth(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		data := &dateOfBirthData{
			App:  appData,
			Lpa:  lpa,
			Form: newDateOfBirthForm(appData.Localizer),
		}

		data.Form.Dob.Set(certificateProvider.DateOfBirth)

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
	forms.Form
	Dob              *forms.Date
	IgnoreDobWarning *forms.String
}

func newDateOfBirthForm(l Localizer) *dateOfBirthForm {
	return &dateOfBirthForm{
		Dob: forms.NewDate("date-of-birth", l.T("dateOfBirth")).
			NotEmpty().
			Real().
			Past().
			BeforeYears(18).WithError(forms.ErrorMessage(l.T("youAreUnder18Error"))),
		IgnoreDobWarning: forms.NewString("ignore-dob-warning", ""),
	}
}

func (f *dateOfBirthForm) DobWarning() string {
	var (
		hundredYearsEarlier = date.Today().AddDate(-100, 0, 0)
	)

	if !f.Dob.Value.IsZero() && f.Dob.Value.Before(hundredYearsEarlier) {
		return "dateOfBirthIsOver100"
	}

	return ""
}

func (f *dateOfBirthForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.Dob, f.IgnoreDobWarning)
}
