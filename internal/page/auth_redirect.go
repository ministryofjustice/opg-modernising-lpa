package page

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

func AuthRedirect(logger Logger, sessionStore SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oneLoginSession, err := sessionStore.OneLogin(r)
		if err != nil {
			logger.InfoContext(r.Context(), "problem retrieving onelogin session", slog.Any("err", err))
			return
		}

		if oneLoginSession.State != r.FormValue("state") {
			logger.InfoContext(r.Context(), "state incorrect")
			return
		}

		lang := localize.En
		if oneLoginSession.Locale == "cy" {
			lang = localize.Cy
		}

		appData := appcontext.Data{Lang: lang, LpaID: oneLoginSession.LpaID}

		appData.Redirect(w, r, oneLoginSession.Redirect+"?"+r.URL.RawQuery)
	}
}
