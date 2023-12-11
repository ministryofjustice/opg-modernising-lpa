package page

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

func LoginCallback(oneLoginClient LoginCallbackOneLoginClient, sessionStore sesh.Store, redirect Path, dashboardStore DashboardStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
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

		log.Printf("checking sub '%s' exists", userInfo.Sub)

		exists, err := dashboardStore.SubExists(r.Context(), base64.StdEncoding.EncodeToString([]byte(userInfo.Sub)))
		if err != nil {
			return err
		}

		if exists {
			redirect = Paths.Dashboard
		}

		return appData.Redirect(w, r, redirect.Format())
	}
}
