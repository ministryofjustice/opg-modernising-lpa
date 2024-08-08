package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type mobileNumberData struct {
	App    appcontext.Data
	Donor  *lpadata.Lpa
	Form   *mobileNumberForm
	Errors validation.List
}

type mobileNumberForm struct {
	Mobile string
}

func MobileNumber(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		data := &mobileNumberData{
			App: appData,
			Form: &mobileNumberForm{
				Mobile: attorneyProvidedDetails.Mobile,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readMobileNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.Mobile = data.Form.Mobile
				if attorneyProvidedDetails.Tasks.ConfirmYourDetails == task.StateNotStarted {
					attorneyProvidedDetails.Tasks.ConfirmYourDetails = task.StateInProgress
				}

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				return attorney.PathYourPreferredLanguage.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

func readMobileNumberForm(r *http.Request) *mobileNumberForm {
	return &mobileNumberForm{
		Mobile: page.PostFormString(r, "mobile"),
	}
}

func (f *mobileNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("mobile", "mobile", f.Mobile,
		validation.Mobile())

	return errors
}
