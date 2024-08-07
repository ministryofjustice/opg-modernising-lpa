package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysGuidanceData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func ChooseAttorneysGuidance(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &chooseAttorneysGuidanceData{
			App:   appData,
			Donor: provided,
		}

		if r.Method == http.MethodPost {
			return donor.PathChooseAttorneys.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
		}

		return tmpl(w, data)
	}
}
