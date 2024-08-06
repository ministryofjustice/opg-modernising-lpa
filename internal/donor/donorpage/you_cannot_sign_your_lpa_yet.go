package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

func YouCannotSignYourLpaYet(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.Under18ActorDetails()) == 0 {
			return appData.Redirect(w, r, donor.PathTaskList.Format(provided.LpaID))
		}

		return tmpl(w, guidanceData{App: appData, Donor: provided})
	}
}
