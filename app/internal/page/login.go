package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type loginClient interface {
	AuthCodeURL(state, nonce string) string
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

		authCodeURL := c.AuthCodeURL(state, nonce)

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
