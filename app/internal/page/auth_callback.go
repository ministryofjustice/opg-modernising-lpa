package page

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

type authCallbackClient interface {
	Exchange(string) (string, error)
	UserInfo(string) (signin.UserInfo, error)
}

func AuthCallback(c authCallbackClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt, err := c.Exchange(r.FormValue("code"))
		if err != nil {
			log.Fatalf("Error getting token: %v", err)
		}

		userInfo, err := c.UserInfo(jwt)
		if err != nil {
			log.Fatalf("Error getting user info: %v", err)
		}

		q := url.Values{
			"email": {userInfo.Email},
		}

		http.Redirect(w, r, "/home?"+q.Encode(), http.StatusFound)
	}
}
