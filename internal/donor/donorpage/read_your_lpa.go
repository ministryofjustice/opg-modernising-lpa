package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readYourLpaData struct {
	App      appcontext.Data
	FixedApp appcontext.Data
	Errors   validation.List
	Donor    *donordata.Provided
}

func ReadYourLpa(tmpl template.Template, bundle Bundle) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		fixedAppData := appData
		fixedAppData.Lang = donor.Donor.LpaLanguagePreference
		fixedAppData.Localizer = bundle.For(donor.Donor.LpaLanguagePreference)

		data := &readYourLpaData{
			App:      appData,
			FixedApp: fixedAppData,
			Donor:    donor,
		}

		return tmpl(w, data)
	}
}
