package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func Login(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, secure bool, randomString func(int) string) http.HandlerFunc {
	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   10 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   secure,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		locale := "en"

		if r.URL.Query().Has("locale") {
			locale = r.URL.Query().Get("locale")
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, false)

		params := sessions.NewSession(store, "params")
		params.Values = map[interface{}]interface{}{
			"state":  state,
			"nonce":  nonce,
			"locale": locale,
		}
		params.Options = cookieOptions

		if err := store.Save(r, w, params); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
