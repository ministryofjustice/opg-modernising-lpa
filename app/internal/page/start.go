package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type startData struct {
	Page             string
	L                localize.Localizer
	Lang             Lang
	CookieConsentSet bool
}

func Start(logger Logger, localizer localize.Localizer, lang Lang, tmpl template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &startData{
			Page:             startPath,
			L:                localizer,
			Lang:             lang,
			CookieConsentSet: cookieConsentSet(r),
		}

		if err := tmpl(w, data); err != nil {
			logger.Print(err)
		}
	}
}
