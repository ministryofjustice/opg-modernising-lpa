package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
)

func SignOut(logger Logger, sessionStore sesh.Store, oneLoginClient OneLoginClient, appPublicURL string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var idToken string
		if session, err := sesh.Login(sessionStore, r); err == nil && session != nil {
			idToken = session.IDToken
		}

		if err := sesh.ClearLoginSession(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire session: %s", err.Error()))
		}

		http.Redirect(w, r, oneLoginClient.EndSessionURL(idToken, appPublicURL+Paths.Start), http.StatusFound)
		return nil
	}
}
