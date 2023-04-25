package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaProgressData struct {
	App                 page.AppData
	Lpa                 *page.Lpa
	CertificateProvider *actor.CertificateProvider
	Errors              validation.List
}

func LpaProgress(tmpl template.Template, lpaStore LpaStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: lpa.ID})

		certificateProvider, err := certificateProviderStore.Get(ctx)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
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
