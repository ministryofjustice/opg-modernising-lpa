package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    appcontext.Data
	Errors validation.List
}

func Guidance(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &guidanceData{
			App: appData,
		}

		return tmpl(w, data)
	}
}
