package page

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
)

type homeData struct {
	UserEmail string
}

func Home(tmpl template.Template, appBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestURI, err := url.Parse(r.RequestURI)

		if err != nil {
			log.Fatalf("Error parsing requestURI: %v", err)
		}

		userEmail := requestURI.Query().Get("user_email")

		err = tmpl(w, homeData{UserEmail: userEmail})

		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}

		log.Println("home")
	}
}
