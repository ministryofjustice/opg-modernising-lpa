package page

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func AuthRedirect(logger Logger, store sesh.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oneLoginSession, err := sesh.OneLogin(store, r)
		if err != nil {
			logger.Info("problem retrieving onelogin session", slog.Any("err", err))
			return
		}

		if oneLoginSession.State != r.FormValue("state") {
			logger.Info("state incorrect")
			return
		}

		lang := localize.En
		if oneLoginSession.Locale == "cy" {
			lang = localize.Cy
		}

		appData := AppData{Lang: lang, LpaID: oneLoginSession.LpaID}

		appData.Redirect(w, r, oneLoginSession.Redirect+"?"+r.URL.RawQuery)
	}
}
