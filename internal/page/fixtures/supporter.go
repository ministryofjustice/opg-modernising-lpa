package fixtures

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OrganisationStore interface {
	Create(context.Context, string) error
}

func Supporter(sessionStore sesh.Store, organisationStore OrganisationStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			organisation = r.FormValue("organisation")
			redirect     = r.FormValue("redirect")

			supporterSub       = random.String(16)
			supporterSessionID = base64.StdEncoding.EncodeToString([]byte(supporterSub))
			ctx                = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: supporterSessionID})
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: supporterSub, Email: testEmail}); err != nil {
			return err
		}

		if organisation == "1" {
			if err := organisationStore.Create(ctx, random.String(12)); err != nil {
				return err
			}
		}

		if redirect != page.Paths.Supporter.EnterOrganisationName.Format() {
			redirect = "/supporter/" + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
