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
	Donor    *actor.DonorProvidedDetails
	Progress page.Progress
	Errors   validation.List
}

func LpaProgress(tmpl template.Template, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore, progressTracker ProgressTracker) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		certificateProvider, err := certificateProviderStore.GetAny(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		if certificateProvider == nil {
			certificateProvider = &actor.CertificateProviderProvidedDetails{}
		}

		attorneys, err := attorneyStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		data := &lpaProgressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(donor, certificateProvider, attorneys),
		}

		return tmpl(w, data)
	}
}
