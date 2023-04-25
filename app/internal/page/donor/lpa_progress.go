package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type lpaProgressData struct {
	App                 page.AppData
	Lpa                 *page.Lpa
	CertificateProvider *actor.CertificateProvider
}

func LpaProgress(tmpl template.Template, lpaStore LpaStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &lpaProgressData{
			App:                 appData,
			Lpa:                 lpa,
			CertificateProvider: certificateProvider,
		}

		return tmpl(w, data)
	}
}
