package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoIsEligibleData struct {
	App    page.AppData
	Lpa    *page.Lpa
	Errors validation.List
}

func WhoIsEligible(tmpl template.Template, lpaStore LpaStore, store sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		sc, err := sesh.ShareCode(store, r)
		if err != nil {
			return err
		}

		lpa, err := lpaStore.Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: sc.LpaID}))
		if err != nil {
			return err
		}

		return tmpl(w, whoIsEligibleData{Lpa: lpa, App: appData})
	}
}
