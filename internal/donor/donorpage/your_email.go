package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourEmailData struct {
	App         page.AppData
	Errors      validation.List
	Form        *yourEmailForm
	CanTaskList bool
}

func YourEmail(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &yourEmailData{
			App: appData,
			Form: &yourEmailForm{
				Email: donor.Donor.Email,
			},
			CanTaskList: !donor.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourEmailForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Donor.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CanYouSignYourLpa.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type yourEmailForm struct {
	Email string
}

func readYourEmailForm(r *http.Request) *yourEmailForm {
	return &yourEmailForm{Email: page.PostFormString(r, "email")}
}

func (f *yourEmailForm) Validate() validation.List {
	var errors validation.List

	errors.String("email", "email", f.Email,
		validation.Email())

	return errors
}
