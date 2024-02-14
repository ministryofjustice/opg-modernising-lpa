package page

import (
	"log/slog"
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
			logger.Info("unable to expire session", slog.Any("err", err))
		}

		endSessionURL, err := oneLoginClient.EndSessionURL(idToken, redirectURL)
		if err != nil {
			logger.Info("unable to end onelogin session", slog.Any("err", err))
			endSessionURL = redirectURL
		}

		http.Redirect(w, r, endSessionURL, http.StatusFound)
		return nil
	}
}
