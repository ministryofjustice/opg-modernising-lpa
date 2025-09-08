package donorpage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func EnterAccessCode(logger Logger, donorStore DonorStore, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, accessCode accesscodedata.Link) error {
		if err := donorStore.Link(r.Context(), accessCode, session.Email); err != nil {
			return fmt.Errorf("link donor: %w", err)
		}

		logger.InfoContext(r.Context(), "donor access added", slog.String("lpa_id", accessCode.LpaKey.ID()))

		if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineDonor); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return page.PathDashboard.Redirect(w, r, appData)
	}
}
