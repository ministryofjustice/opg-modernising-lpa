package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                     page.AppData
	Errors                  validation.List
	Lpa                     *lpastore.Lpa
	Attorney                lpastore.Attorney
	TrustCorporation        lpastore.TrustCorporation
	AttorneyProvidedDetails *attorneydata.Provided
}

func ConfirmYourDetails(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		if r.Method == http.MethodPost {
			attorneyProvidedDetails.Tasks.ConfirmYourDetails = task.StateCompleted

			if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
				return err
			}

			return page.Paths.Attorney.TaskList.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
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
