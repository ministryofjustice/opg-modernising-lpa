package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourEmailData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *yourEmailForm
	CanTaskList bool
}

func YourEmail(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourEmailData{
			App: appData,
			Form: &yourEmailForm{
				Email: provided.Donor.Email,
			},
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourEmailForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathCanYouSignYourLpa.Redirect(w, r, appData, provided)
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
