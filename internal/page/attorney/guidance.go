package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    page.AppData
	Errors validation.List
	Donor  *actor.DonorProvidedDetails
}

func Guidance(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, _ *actor.AttorneyProvidedDetails) error {
		data := &guidanceData{
			App: appData,
		}

		if donorStore != nil {
			donor, err := donorStore.GetAny(r.Context())
			if err != nil {
				return err
			}
			data.Donor = donor
		}

		return tmpl(w, data)
	}
}
