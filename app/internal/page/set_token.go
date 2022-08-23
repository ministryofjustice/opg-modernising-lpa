package page

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	govuksignin "github.com/ministryofjustice/opg-modernising-lpa/internal/gov_uk_sign_in"
)

func SetToken(c govuksignin.Client, appBaseURL, clientID, JTI string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("/auth/callback")

		jwt, err := c.GetToken(fmt.Sprintf("%s:%s", appBaseURL, "/home"), clientID, JTI)

		if err != nil {
			log.Fatalf("Error getting token: %v", err)
		}

		userInfo, err := c.GetUserInfo(jwt)

		if err != nil {
			log.Fatalf("Error getting user info: %v", err)
		}

		redirectURL, err := url.Parse(fmt.Sprintf("%s/home", appBaseURL))

		if err != nil {
			log.Fatalf("Error parsing redirect URL: %v", err)
		}

		redirectURL.Query().Add("email", userInfo.Email)

		log.Printf("redirecting to %s", redirectURL.String())

		http.Redirect(w, r, redirectURL.String(), http.StatusPermanentRedirect)
	}
}
