package page

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
)

type homeData struct {
	UserEmail string
	SignInURL string
}

func Home(tmpl template.Template, signInURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestURI, err := url.Parse(r.RequestURI)

		if err != nil {
			log.Fatalf("Error parsing requestURI: %v", err)
		}

		userEmail := requestURI.Query().Get("email")

		data := homeData{UserEmail: userEmail, SignInURL: signInURL}
		err = tmpl(w, data)

		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
}
