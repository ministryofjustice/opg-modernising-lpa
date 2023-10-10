package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type progressData struct {
	App             page.AppData
	Errors          validation.List
	Lpa             *page.Lpa
	Signed          bool
	AttorneysSigned bool
}

func Progress(tmpl template.Template, attorneyStore AttorneyStore, donorStore DonorStore, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		attorneys, err := attorneyStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		progress := lpa.Progress(certificateProvider, attorneys)

		data := &progressData{
			App:             appData,
			Lpa:             lpa,
			Signed:          attorneyProvidedDetails.Signed(certificateProvider.Certificate.Agreed),
			AttorneysSigned: progress.AttorneysSigned.Completed(),
		}

		return tmpl(w, data)
	}
}
