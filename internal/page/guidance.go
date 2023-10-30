package page

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    AppData
	Query  url.Values
	Errors validation.List
}

func Guidance(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &guidanceData{
			App:   appData,
			Query: r.URL.Query(),
		}

		return tmpl(w, data)
	}
}
