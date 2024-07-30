package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func YouCannotSignYourLpaYet(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if len(donor.Under18ActorDetails()) == 0 {
			return appData.Redirect(w, r, page.Paths.TaskList.Format(donor.LpaID))
		}

		return tmpl(w, guidanceData{App: appData, Donor: donor})
	}
}
