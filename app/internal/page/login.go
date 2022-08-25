package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

func Login(c signin.Client, clientID, redirectURL, signInPublicURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authCodeURL := c.AuthCodeURL(redirectURL, clientID, "state-value", "nonce-value", "scope-value", signInPublicURL)

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
