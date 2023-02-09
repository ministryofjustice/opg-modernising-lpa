package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func CertificateProviderLogin(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		locale := ""
		if appData.Lang == Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, true)

		sessionID := r.FormValue("sessionId")
		lpaID := r.FormValue("lpaId")

		if err := setOneLoginSession(store, r, w, &OneLoginSession{
			State:               state,
			Nonce:               nonce,
			Locale:              locale,
			CertificateProvider: true,
			Identity:            true,
			SessionID:           sessionID,
			LpaID:               lpaID,
		}); err != nil {
			logger.Print(err)
			return nil
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
