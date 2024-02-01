package supporter

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

func LoginCallback(oneLoginClient LoginCallbackOneLoginClient, sessionStore sesh.Store) page.Handler {
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

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{
			IDToken: idToken,
			Sub:     userInfo.Sub,
			Email:   userInfo.Email,
		}); err != nil {
			return err
		}

		return appData.Redirect(w, r, page.Paths.Supporter.EnterOrganisationName.Format())
	}
}
