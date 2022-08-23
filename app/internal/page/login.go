package page

import (
	"fmt"
	"log"
	"net/http"

	govuksignin "github.com/ministryofjustice/opg-modernising-lpa/internal/gov_uk_sign_in"
)

func Login(c govuksignin.Client, appBaseURL, clientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/login")

		redirectURL := fmt.Sprintf("%s%s", appBaseURL, c.AuthCallbackPath)
		err := c.AuthorizeAndRedirect(redirectURL, clientID, "state-value", "nonce-value", "scope-value")

		if err != nil {
			log.Fatalf("Error GETting authorize: %v", err)
		}
	}
}
