package donor

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysGuidanceData struct {
	App    page.AppData
	Errors validation.List
	Donor  *actor.DonorProvidedDetails
}

func ChooseAttorneysGuidance(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &chooseAttorneysGuidanceData{
			App:   appData,
			Donor: donor,
		}

		if r.Method == http.MethodPost {
			return page.Paths.ChooseAttorneys.RedirectQuery(w, r, appData, donor, url.Values{"id": {newUID().String()}})
		}

		return tmpl(w, data)
	}
}
