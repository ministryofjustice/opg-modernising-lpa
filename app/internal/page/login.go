package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type loginClient interface {
	AuthCodeURL(state, nonce, locale string) string
}

func Login(logger Logger, c loginClient, store sessions.Store, secure bool, randomString func(int) string) http.HandlerFunc {
	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   10 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   secure,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		state := randomString(12)
		nonce := randomString(12)
		locale := "en"

		if r.URL.Query().Has("locale") {
			locale = r.URL.Query().Get("locale")
		}

		authCodeURL := c.AuthCodeURL(state, nonce, locale)

		params := sessions.NewSession(store, "params")
		params.Values = map[interface{}]interface{}{
			"state": state,
			"nonce": nonce,
		}
		params.Options = cookieOptions

		if err := store.Save(r, w, params); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
