package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

func MakeOrAddAnLPA(tmpl template.Template, donorStore DonorStore, dashboardStore DashboardStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		results, err := dashboardStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		data := makeOrAddAnLPAData{
			App:          appData,
			HasDonorLPAs: len(results.ByActorType(actor.TypeDonor)) > 0,
		}

		if r.Method == http.MethodPost {
			provided, err := donorStore.Create(r.Context())
			if err != nil {
				return err
			}

			path := donor.PathYourName
			if data.HasDonorLPAs {
				path = donor.PathMakeANewLPA
			}

			if err := eventClient.SendMetric(r.Context(), provided.LpaID, event.CategoryFunnelStartRate, event.MeasureOnlineDonor); err != nil {
				return fmt.Errorf("sending metric: %w", err)
			}

			return path.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}

type makeOrAddAnLPAData struct {
	App          appcontext.Data
	Errors       validation.List
	HasDonorLPAs bool
}
