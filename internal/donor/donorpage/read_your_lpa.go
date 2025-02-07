package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readYourLpaData struct {
	App            appcontext.Data
	LpaLanguageApp appcontext.Data
	Errors         validation.List
	Donor          *donordata.Provided
}

func ReadYourLpa(tmpl template.Template, bundle Bundle) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpaLanguageApp := appData
		lpaLanguageApp.Lang = donor.Donor.LpaLanguagePreference
		lpaLanguageApp.Localizer = bundle.For(donor.Donor.LpaLanguagePreference)

		data := &readYourLpaData{
			App:            appData,
			LpaLanguageApp: lpaLanguageApp,
			Donor:          donor,
		}

		return tmpl(w, data)
	}
}
