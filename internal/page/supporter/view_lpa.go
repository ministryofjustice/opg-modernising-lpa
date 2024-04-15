package supporter

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type viewLPAData struct {
	App      page.AppData
	Errors   validation.List
	Lpa      *lpastore.Lpa
	Progress page.Progress
}

func ViewLPA(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore, progressTracker ProgressTracker) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.GetAny(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		return tmpl(w, &viewLPAData{
			App:      appData,
			Lpa:      lpa,
			Progress: progressTracker.Progress(lpa, certificateProvider),
		})
	}
}
