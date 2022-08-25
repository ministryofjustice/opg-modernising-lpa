package page

import (
	"log"
	"net/http"

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
		data := homeData{
			UserEmail: r.FormValue("email"),
			SignInURL: signInURL,
			L:         localizer,
			Lang:      lang,
		}

		err := tmpl(w, data)
		if err != nil {
			log.Fatalf("Error rendering template: %v", err)
		}
	}
}
