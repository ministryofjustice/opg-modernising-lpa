package page

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

func SignOut(logger Logger, sessionStore SessionStore, oneLoginClient OneLoginClient, donorStartURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		var idToken string
		if session, err := sessionStore.Login(r); err == nil && session != nil {
			idToken = session.IDToken
		}

		if err := sessionStore.ClearLogin(r, w); err != nil {
			logger.InfoContext(r.Context(), "unable to expire session", slog.Any("err", err))
		}

		endSessionURL, err := oneLoginClient.EndSessionURL(idToken, donorStartURL)
		if err != nil {
			logger.InfoContext(r.Context(), "unable to end onelogin session", slog.Any("err", err))
			endSessionURL = donorStartURL
		}

		logger.InfoContext(r.Context(), "logout")

		http.Redirect(w, r, endSessionURL, http.StatusFound)
		return nil
	}
}
