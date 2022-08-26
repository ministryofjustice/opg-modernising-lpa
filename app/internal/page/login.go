package page

import (
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

func Login(c *signin.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authCodeURL := c.AuthCodeURL("state-value", "nonce-value", "scope-value")

		log.Println("hey", authCodeURL)
		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
