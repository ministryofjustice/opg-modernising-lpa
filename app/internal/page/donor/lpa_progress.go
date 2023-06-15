package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type lpaProgressData struct {
	App                 page.AppData
	Lpa                 *page.Lpa
	CertificateProvider *actor.CertificateProviderProvidedDetails
	Errors              validation.List
}

func LpaProgress(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: lpa.ID})

		certificateProvider, err := certificateProviderStore.GetAny(ctx)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		if certificateProvider == nil {
			certificateProvider = &actor.CertificateProviderProvidedDetails{}
		}

		data := &lpaProgressData{
			App:                 appData,
			Lpa:                 lpa,
			CertificateProvider: certificateProvider,
		}

		return tmpl(w, data)
	}
}
