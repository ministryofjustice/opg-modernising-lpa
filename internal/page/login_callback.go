package page

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
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

func LoginCallback(logger Logger, oneLoginClient LoginCallbackOneLoginClient, sessionStore LoginCallbackSessionStore, redirect Path, dashboardStore DashboardStore, actorType actor.Type) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		if error := r.FormValue("error"); error != "" {
			logger.InfoContext(r.Context(), "login error",
				slog.String("error", error),
				slog.String("error_description", r.FormValue("error_description")))
			return errors.New("access denied")
		}

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

		results, err := dashboardStore.GetAll(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: session.SessionID()}))
		if err != nil {
			return err
		}

		if !results.Empty() {
			session.HasLPAs = true
		}

		if err := sessionStore.SetLogin(r, w, session); err != nil {
			return err
		}

		logger.InfoContext(r.Context(), "login", slog.String("session_id", session.SessionID()))

		finalRedirect := redirect

		if actorType.IsDonor() {
			if len(results.Donor) == 0 {
				finalRedirect = PathMakeOrAddAnLPA
			}
		} else {
			if len(results.ByActorType(actorType)) > 0 {
				finalRedirect = PathDashboard
			}
		}

		return appData.Redirect(w, r, finalRedirect.Format())
	}
}
