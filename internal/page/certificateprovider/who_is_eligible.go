package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoIsEligibleData struct {
	App             page.AppData
	DonorFullName   string
	DonorFirstNames string
	Errors          validation.List
}

func WhoIsEligible(tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, whoIsEligibleData{DonorFullName: lpa.Donor.FullName(), DonorFirstNames: lpa.Donor.FirstNames, App: appData})
	}
}
