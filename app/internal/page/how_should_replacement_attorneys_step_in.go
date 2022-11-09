package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howShouldReplacementAttorneysStepInData struct{}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		// Todo in next ticket

		data := &howShouldReplacementAttorneysStepInData{}

		return tmpl(w, data)
	}
}
