package attorneypage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func EnterAccessCode(attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, accessCode accesscodedata.Link) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("resolving lpa: %w", err)
		}

		lpaAttorney, found := lpa.Attorneys.Get(accessCode.ActorUID)

		if found && lpaAttorney.Channel.IsPaper() && !lpaAttorney.SignedAt.IsZero() {
			if err := lpaStoreClient.SendPaperAttorneyAccessOnline(r.Context(), accessCode.LpaUID, session.Email, accessCode.ActorUID); err != nil {
				return fmt.Errorf("sending attorney email to LPA store: %w", err)
			}

			return page.PathDashboard.Redirect(w, r, appData)
		}

		if _, err := attorneyStore.Create(r.Context(), accessCode, session.Email); err != nil {
			return fmt.Errorf("create attorney: %w", err)
		}

		if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineAttorney); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return attorney.PathCodeOfConduct.Redirect(w, r, appData, appData.LpaID)
	}
}
