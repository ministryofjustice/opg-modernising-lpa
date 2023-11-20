package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type feeDeniedData struct {
	Donor  *actor.DonorProvidedDetails
	Errors validation.List
	App    page.AppData
}

func FeeDenied(tmpl template.Template, payer Payer) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if r.Method == http.MethodPost {
			return payer.Pay(appData, w, r, donor)
		}

		return tmpl(w, feeDeniedData{Donor: donor, App: appData})
	}
}
