package page

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
)

type authRedirectClient interface {
	Exchange(string) (string, error)
	UserInfo(string) (signin.UserInfo, error)
}

func AuthRedirect(logger Logger, c authRedirectClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt, err := c.Exchange(r.FormValue("code"))
		if err != nil {
			logger.Print(err)
			return
		}

		userInfo, err := c.UserInfo(jwt)
		if err != nil {
			logger.Print(err)
			return
		}

		q := url.Values{
			"email": {userInfo.Email},
		}

		http.Redirect(w, r, "/home?"+q.Encode(), http.StatusFound)
	}
}
