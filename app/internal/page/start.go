package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

func Start(tmpl template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl(w, nil)
	}
}
