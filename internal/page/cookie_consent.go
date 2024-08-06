package page

import (
	"net/http"
)

func CookieConsent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		consent := "reject"
		if r.PostFormValue("cookies") == "accept" {
			consent = "accept"
		}

		http.SetCookie(w, &http.Cookie{
			Name:   "cookies-consent",
			Value:  consent,
			MaxAge: 365 * 24 * 60 * 60,
			Path:   "/",
		})

		redirectURL := r.PostFormValue("cookies-redirect")
		if len(redirectURL) <= 1 || redirectURL[0] != '/' || redirectURL[1] == '/' || redirectURL[1] == '\\' {
			redirectURL = PathStart.String()
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
