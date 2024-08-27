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

type yourMobileData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *yourMobileForm
	CanTaskList bool
}

func YourMobile(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourMobileData{
			App: appData,
			Form: &yourMobileForm{
				Mobile: provided.Donor.Mobile,
			},
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourMobileForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.Mobile = data.Form.Mobile

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathReceivingUpdatesAboutYourLpa.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourMobileForm struct {
	Mobile string
}

func readYourMobileForm(r *http.Request) *yourMobileForm {
	return &yourMobileForm{Mobile: page.PostFormString(r, "mobile")}
}

func (f *yourMobileForm) Validate() validation.List {
	var errors validation.List

	errors.String("mobile", "mobile", f.Mobile,
		validation.Mobile())

	return errors
}
