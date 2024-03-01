package page

import (
	"log/slog"
	"net/http"
)

func SignOut(logger Logger, sessionStore SessionStore, oneLoginClient OneLoginClient, appPublicURL string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		redirectURL := appPublicURL + Paths.Start.Format()

		var idToken string
		if session, err := sessionStore.Login(r); err == nil && session != nil {
			idToken = session.IDToken
		}

		if err := sessionStore.ClearLogin(r, w); err != nil {
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
