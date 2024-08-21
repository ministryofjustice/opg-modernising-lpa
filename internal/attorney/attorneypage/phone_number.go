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

type phoneNumberData struct {
	App    appcontext.Data
	Donor  *lpadata.Lpa
	Form   *phoneNumberForm
	Errors validation.List
}

type phoneNumberForm struct {
	Phone string
}

func PhoneNumber(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		data := &phoneNumberData{
			App: appData,
			Form: &phoneNumberForm{
				Phone: attorneyProvidedDetails.Phone,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readPhoneNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.Phone = data.Form.Phone
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

func readPhoneNumberForm(r *http.Request) *phoneNumberForm {
	return &phoneNumberForm{
		Phone: page.PostFormString(r, "phone"),
	}
}

func (f *phoneNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("phone", "phone", f.Phone,
		validation.Phone())

	return errors
}
