package attorneypage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func EnterAccessCode(attorneyStore AttorneyStore, lpaStoreClient LpaStoreClient, eventClient EventClient) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, accessCode accesscodedata.Link) error {
		lpaAttorney, found := lpa.Attorneys.Get(accessCode.ActorUID)

		if found && lpaAttorney.Channel.IsPaper() && !lpaAttorney.SignedAt.IsZero() {
			if err := lpaStoreClient.SendPaperAttorneyAccessOnline(r.Context(), accessCode.LpaUID, session.Email, accessCode.ActorUID); err != nil {
				return fmt.Errorf("sending attorney email to LPA store: %w", err)
			}

			return page.PathDashboard.Redirect(w, r, appData)
		}

		if _, err := attorneyStore.Create(r.Context(), accessCode, session.Email); err != nil {
			return err
		}

		if err := eventClient.SendMetric(r.Context(), lpa.LpaID+"/"+session.Sub, event.CategoryFunnelStartRate, event.MeasureOnlineAttorney); err != nil {
			return fmt.Errorf("sending metric: %w", err)
		}

		return attorney.PathCodeOfConduct.Redirect(w, r, appData, appData.LpaID)
	}
}
