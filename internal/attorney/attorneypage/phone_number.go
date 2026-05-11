package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type phoneNumberData struct {
	App    appcontext.Data
	Donor  *lpadata.Lpa
	Form   *phoneNumberForm
	Errors validation.List
}

func PhoneNumber(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided, lpa *lpadata.Lpa) error {
		_, mobile, _ := lpa.Attorney(provided.UID)
		if provided.PhoneSet {
			mobile = provided.Phone
		}

		data := &phoneNumberData{
			App:  appData,
			Form: newPhoneNumberForm(appData.Localizer),
		}

		data.Form.Phone.Input = mobile

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				provided.Phone = data.Form.Phone.Value
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

type phoneNumberForm struct {
	Phone  *newforms.String
	Errors []newforms.Field
}

func newPhoneNumberForm(l Localizer) *phoneNumberForm {
	return &phoneNumberForm{
		Phone: newforms.NewString("phone", l.T("phone")).
			Phone(),
	}
}

func (f *phoneNumberForm) Parse(r *http.Request) bool {
	f.Errors = newforms.ParsePostForm(r,
		f.Phone,
	)

	return len(f.Errors) == 0
}
