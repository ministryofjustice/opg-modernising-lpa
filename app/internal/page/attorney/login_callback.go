package attorney

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func LoginCallback(oneLoginClient OneLoginClient, sessionStore sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.Attorney || oneLoginSession.Identity {
			return errors.New("attorney callback with incorrect session")
		}

		idToken, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if err := sesh.SetAttorney(sessionStore, r, w, &sesh.AttorneySession{
			IDToken: idToken,
			Sub:     userInfo.Sub,
			Email:   userInfo.Email,
		}); err != nil {
			return err
		}

		return appData.Redirect(w, r, nil, page.Paths.Attorney.EnterReferenceNumber)
	}
}
