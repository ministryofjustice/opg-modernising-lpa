package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type loginClient interface {
	AuthCodeURL(state, nonce string) string
}

func Login(logger Logger, c loginClient, store sessions.Store, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomString(12)

		authCodeURL := c.AuthCodeURL(state, "nonce-value")

		session, err := store.New(r, "params")
		if err != nil {
			logger.Print(err)
			return
		}
		session.Values = map[interface{}]interface{}{"state": state}
		if err := store.Save(r, w, session); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
