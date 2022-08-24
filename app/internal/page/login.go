package page

import (
	"fmt"
	"net/http"

	govuksignin "github.com/ministryofjustice/opg-modernising-lpa/internal/gov_uk_sign_in"
)

func Login(c govuksignin.Client, appPublicURL, clientID, signInBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := fmt.Sprintf("%s%s", appPublicURL, c.AuthCallbackPath)
		c.AuthorizeAndRedirect(w, r, redirectURL, clientID, "state-value", "nonce-value", "scope-value", signInBaseURL)
	}
}
