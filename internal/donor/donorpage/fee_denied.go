package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type feeDeniedData struct {
	Donor  *donordata.Provided
	Errors validation.List
	App    page.AppData
}

func FeeDenied(tmpl template.Template, payer Handler) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if r.Method == http.MethodPost {
			return payer(appData, w, r, donor)
		}

		return tmpl(w, feeDeniedData{Donor: donor, App: appData})
	}
}
