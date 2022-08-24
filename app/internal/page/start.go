package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type startData struct {
	Page string
	L    localize.Localizer
	Lang Lang
}

func Start(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template) http.HandlerFunc {
	data := &startData{
		Page: "/",
		L:    localizer,
		Lang: lang,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}
