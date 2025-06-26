package attorneypage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

func EnterAccessCode(attorneyStore AttorneyStore, lpaStoreClient LpaStoreClient, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, shareCode sharecodedata.Link) error {
		lpaAttorney, found := lpa.Attorneys.Get(shareCode.ActorUID)

		if found && lpaAttorney.Channel.IsPaper() && !lpaAttorney.SignedAt.IsZero() {
			if err := lpaStoreClient.SendPaperAttorneyAccessOnline(r.Context(), shareCode.LpaUID, session.Email, shareCode.ActorUID); err != nil {
				return fmt.Errorf("sending attorney email to LPA store: %w", err)
			}

			return page.PathDashboard.Redirect(w, r, appData)
		}

		if _, err := attorneyStore.Create(r.Context(), shareCode, session.Email); err != nil {
			return err
		}

		if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineAttorney); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return attorney.PathCodeOfConduct.Redirect(w, r, appData, appData.LpaID)
	}
}
