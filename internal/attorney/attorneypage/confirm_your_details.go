package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                     appcontext.Data
	Errors                  validation.List
	Lpa                     *lpadata.Lpa
	Attorney                lpadata.Attorney
	TrustCorporation        lpadata.TrustCorporation
	AttorneyProvidedDetails *attorneydata.Provided
	DonorProvidedMobile     string
}

func ConfirmYourDetails(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided) error {
		if r.Method == http.MethodPost {
			provided.Tasks.ConfirmYourDetails = task.StateCompleted

			if err := attorneyStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return attorney.PathTaskList.Redirect(w, r, appData, provided.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		_, mobile, _ := lpa.Attorney(provided.UID)

		data := &confirmYourDetailsData{
			App:                     appData,
			Lpa:                     lpa,
			AttorneyProvidedDetails: provided,
			DonorProvidedMobile:     mobile,
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		if appData.IsTrustCorporation() {
			data.TrustCorporation = attorneys.TrustCorporation
		} else {
			data.Attorney, _ = attorneys.Get(provided.UID)
		}

		return tmpl(w, data)
	}
}
