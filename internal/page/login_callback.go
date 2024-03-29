package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type LoginCallbackSessionStore interface {
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
}

func LoginCallback(oneLoginClient LoginCallbackOneLoginClient, sessionStore LoginCallbackSessionStore, redirect Path, dashboardStore DashboardStore, actorType actor.Type) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sessionStore.OneLogin(r)
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
			Sub:     userInfo.Sub,
			Email:   userInfo.Email,
		}

		if err := sessionStore.SetLogin(r, w, session); err != nil {
			return err
		}

		if actorType != actor.TypeDonor {
			exists, err := dashboardStore.SubExistsForActorType(r.Context(), session.SessionID(), actorType)

			if err != nil {
				return err
			}

			if exists {
				redirect = Paths.Dashboard
			}
		}

		return appData.Redirect(w, r, redirect.Format())
	}
}
