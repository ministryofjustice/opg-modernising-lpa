package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func Login(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, secure bool, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := "en"

		if r.URL.Query().Has("locale") {
			locale = r.URL.Query().Get("locale")
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, false)

		if err := setOneLoginSession(store, r, w, &OneLoginSession{
			State:  state,
			Nonce:  nonce,
			Locale: locale,
		}); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
