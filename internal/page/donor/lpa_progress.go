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
	App      page.AppData
	Lpa      *page.Lpa
	Progress page.Progress
	Errors   validation.List
}

func LpaProgress(tmpl template.Template, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: lpa.ID})

		certificateProvider, err := certificateProviderStore.GetAny(ctx)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		if certificateProvider == nil {
			certificateProvider = &actor.CertificateProviderProvidedDetails{}
		}

		attorneys, err := attorneyStore.GetAny(ctx)
		if err != nil {
			return err
		}

		data := &lpaProgressData{
			App:      appData,
			Lpa:      lpa,
			Progress: lpa.Progress(certificateProvider, attorneys),
		}

		return tmpl(w, data)
	}
}
