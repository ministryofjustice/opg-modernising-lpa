package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type startData struct {
	L    localize.Localizer
	Lang Lang
}

func Start(logger *logging.Logger, localizer localize.Localizer, lang Lang, tmpl template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl(w, startData{L: localizer, Lang: lang}); err != nil {
			logger.Print(err)
		}
	}
}
