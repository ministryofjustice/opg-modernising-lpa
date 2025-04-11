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

type doYouLiveInTheUKData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func DoYouLiveInTheUK(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &doYouLiveInTheUKData{
			App:   appData,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
			Donor: provided,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfYouLiveInTheUK")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectPath := donor.PathYourAddress
				if data.Form.YesNo.IsNo() {
					redirectPath = donor.PathWhatCountryDoYouLiveIn
				}

				return redirectPath.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
