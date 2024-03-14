package supporter

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type viewLPAData struct {
	App    page.AppData
	Errors validation.List
	Donor  *actor.DonorProvidedDetails
}

func ViewLPA(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		sessionData, err := page.SessionDataFromContext(r.Context())
		if err != nil {
			return err
		}

		lpaID := r.FormValue("id")
		if lpaID == "" {
			return errors.New("lpaID missing from query")
		}

		sessionData.LpaID = lpaID

		ctx := page.ContextWithSessionData(r.Context(), sessionData)

		donor, err := donorStore.Get(ctx)
		if err != nil {
			return err
		}

		return tmpl(w, &viewLPAData{
			App:   appData,
			Donor: donor,
		})
	}
}
