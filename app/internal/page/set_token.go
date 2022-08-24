package page

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

func SetToken(c signin.Client, appPublicURL, clientID, JTI string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt, err := c.GetToken(fmt.Sprintf("%s:%s", appPublicURL, "/home"), clientID, JTI, r.FormValue("code"))

		if err != nil {
			log.Fatalf("Error getting token: %v", err)
		}

		userInfo, err := c.GetUserInfo(jwt)

		if err != nil {
			log.Fatalf("Error getting user info: %v", err)
		}

		redirectURL, err := url.Parse(fmt.Sprintf("%s/home", appPublicURL))

		if err != nil {
			log.Fatalf("Error parsing redirect URL: %v", err)
		}

		q := redirectURL.Query()
		q.Add("email", userInfo.Email)
		redirectURL.RawQuery = q.Encode()

		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	}
}
