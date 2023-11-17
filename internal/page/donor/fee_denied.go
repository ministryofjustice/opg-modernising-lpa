package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type feeDeniedData struct {
	Lpa    *actor.Lpa
	Errors validation.List
	App    page.AppData
}

func FeeDenied(tmpl template.Template, payer Payer) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
		if r.Method == http.MethodPost {
			return payer.Pay(appData, w, r, lpa)
		}

		return tmpl(w, feeDeniedData{Lpa: lpa, App: appData})
	}
}
