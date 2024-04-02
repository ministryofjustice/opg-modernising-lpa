package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                     page.AppData
	Errors                  validation.List
	Lpa                     *lpastore.ResolvedLpa
	Attorney                actor.Attorney
	TrustCorporation        actor.TrustCorporation
	AttorneyProvidedDetails *actor.AttorneyProvidedDetails
}

func ConfirmYourDetails(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		if r.Method == http.MethodPost {
			attorneyProvidedDetails.Tasks.ConfirmYourDetails = actor.TaskCompleted

			if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
				return err
			}

			return page.Paths.Attorney.ReadTheLpa.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &confirmYourDetailsData{
			App:                     appData,
			Lpa:                     lpa,
			AttorneyProvidedDetails: attorneyProvidedDetails,
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		if appData.IsTrustCorporation() {
			data.TrustCorporation = attorneys.TrustCorporation
		} else {
			data.Attorney, _ = attorneys.Get(attorneyProvidedDetails.UID)
		}

		return tmpl(w, data)
	}
}
