package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type homeData struct {
	Page      string
	UserEmail string
	SignInURL string
	L         localize.Localizer
	Lang      Lang
}

func Home(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, signInURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := homeData{
			Page:      homePath,
			UserEmail: r.FormValue("email"),
			SignInURL: signInURL,
			L:         localizer,
			Lang:      lang,
		}

		err := tmpl(w, data)
		if err != nil {
			logger.Print("Error rendering template: %v", err)
		}
	}
}
