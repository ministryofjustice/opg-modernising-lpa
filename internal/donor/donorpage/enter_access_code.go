package donorpage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func EnterAccessCode(logger Logger, donorStore DonorStore, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, link accesscodedata.Link) error {
		if err := donorStore.Link(r.Context(), link, session.Email); err != nil {
			return err
		}

		logger.InfoContext(r.Context(), "donor access added", slog.String("lpa_id", link.LpaKey.ID()))

		if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineDonor); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return page.PathDashboard.Redirect(w, r, appData)
	}
}
