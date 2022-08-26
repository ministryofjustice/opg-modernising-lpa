package page

import (
	"net/http"
)

type loginClient interface {
	AuthCodeURL(state, nonce string) string
}

func Login(c loginClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authCodeURL := c.AuthCodeURL("state-value", "nonce-value")

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
