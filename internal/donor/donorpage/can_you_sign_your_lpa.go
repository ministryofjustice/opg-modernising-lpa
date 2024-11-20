package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type canYouSignYourLpaData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *form.SelectForm[donordata.YesNoMaybe, donordata.YesNoMaybeOptions, *donordata.YesNoMaybe]
	CanTaskList bool
}

func CanYouSignYourLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &canYouSignYourLpaData{
			App:         appData,
			Form:        form.NewSelectForm(provided.Donor.ThinksCanSign, donordata.YesNoMaybeValues, "yesIfCanSign"),
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.ThinksCanSign = data.Form.Selected

				if provided.Donor.ThinksCanSign.IsYes() {
					provided.Donor.CanSign = form.Yes
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.Donor.ThinksCanSign.IsYes() {
					return donor.PathYourPreferredLanguage.Redirect(w, r, appData, provided)
				}

				return donor.PathCheckYouCanSign.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
