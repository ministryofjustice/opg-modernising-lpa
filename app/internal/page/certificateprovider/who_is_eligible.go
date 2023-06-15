package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type whoIsEligibleData struct {
	App             page.AppData
	DonorFullName   string
	DonorFirstNames string
	Errors          validation.List
}

func WhoIsEligible(tmpl template.Template, store sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		sc, err := sesh.ShareCode(store, r)
		if err != nil {
			return err
		}

		return tmpl(w, whoIsEligibleData{DonorFullName: sc.DonorFullName, DonorFirstNames: sc.DonorFirstNames, App: appData})
	}
}
