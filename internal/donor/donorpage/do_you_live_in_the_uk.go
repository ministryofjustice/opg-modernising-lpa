package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type doYouLiveInTheUKData struct {
	App   appcontext.Data
	Form  *newforms.YesNoForm
	Donor *donordata.Provided
}

func DoYouLiveInTheUK(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &doYouLiveInTheUKData{
			App:   appData,
			Form:  newforms.NewYesNoForm(appData.Localizer.T("yesIfYouLiveInTheUK")),
			Donor: provided,
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			redirectPath := donor.PathYourAddress
			if data.Form.YesNo.Value.IsNo() {
				redirectPath = donor.PathWhatCountryDoYouLiveIn
			}

			return redirectPath.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
