package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type homeData struct {
	Page      string
	L         localize.Localizer
	Lang      Lang
	UserEmail string
	SignInURL string
}

func Home(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template, signInURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &homeData{
			Page:      homePath,
			UserEmail: r.FormValue("email"),
			SignInURL: signInURL,
			L:         localizer,
			Lang:      lang,
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}
