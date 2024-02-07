package supporter

import (
	"context"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

func LoginCallback(oneLoginClient LoginCallbackOneLoginClient, sessionStore sesh.Store, organisationStore OrganisationStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}

		idToken, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		session := &sesh.LoginSession{
			IDToken: idToken,
			Sub:     "supporter-" + userInfo.Sub,
			Email:   userInfo.Email,
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: session.SessionID()})

		organisation, err := organisationStore.Get(ctx)
		if err == nil {
			session.OrganisationID = organisation.ID
			session.OrganisationName = organisation.Name
			if err := sesh.SetLoginSession(sessionStore, r, w, session); err != nil {
				return err
			}

			return page.Paths.Supporter.Dashboard.Redirect(w, r, appData)
		}

		if errors.Is(err, dynamo.NotFoundError{}) {
			if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{
				IDToken: idToken,
				Sub:     "supporter-" + userInfo.Sub,
				Email:   userInfo.Email,
			}); err != nil {
				return err
			}
		} else {
			return err
		}

		return page.Paths.Supporter.EnterOrganisationName.Redirect(w, r, appData)
	}
}
