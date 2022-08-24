package page

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"

	"github.com/ministryofjustice/opg-go-common/template"
)

type homeData struct {
	UserEmail string
	SignInURL string
	L         localize.Localizer
	Lang      Lang
}

func Home(tmpl template.Template, signInURL string, localizer localize.Localizer, lang Lang) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestURI, err := url.Parse(r.RequestURI)

		if err != nil {
			log.Fatalf("Error parsing requestURI: %v", err)
		}

		userEmail := requestURI.Query().Get("email")

		data := homeData{
			UserEmail: userEmail,
			SignInURL: signInURL,
			L:         localizer,
			Lang:      lang,
		}
		err = tmpl(w, data)

		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
}
