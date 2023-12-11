package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func SignOut(logger Logger, sessionStore sesh.Store, oneLoginClient OneLoginClient, appPublicURL string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		redirectURL := appPublicURL + Paths.Start.Format()

		var idToken string
		if session, err := sesh.Login(sessionStore, r); err == nil && session != nil {
			idToken = session.IDToken
		}

		if err := sesh.ClearLoginSession(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire session: %s", err.Error()))
		}

		endSessionURL, err := oneLoginClient.EndSessionURL(idToken, redirectURL)
		if err != nil {
			logger.Print(fmt.Sprintf("unable to end onelogin session: %s", err.Error()))
			endSessionURL = redirectURL
		}

		http.Redirect(w, r, endSessionURL, http.StatusFound)
		return nil
	}
}
