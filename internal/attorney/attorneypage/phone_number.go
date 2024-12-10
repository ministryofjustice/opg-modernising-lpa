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

func PhoneNumber(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		_, mobile, _ := lpa.Attorney(provided.UID)
		if provided.PhoneSet {
			mobile = provided.Phone
		}

		data := &phoneNumberData{
			App: appData,
			Form: &phoneNumberForm{
				Phone: mobile,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readPhoneNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Phone = data.Form.Phone
				provided.PhoneSet = true
				if provided.Tasks.ConfirmYourDetails == task.StateNotStarted {
					provided.Tasks.ConfirmYourDetails = task.StateInProgress
				}

				if err := attorneyStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return attorney.PathYourPreferredLanguage.Redirect(w, r, appData, provided.LpaID)
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
