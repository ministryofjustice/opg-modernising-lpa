package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    page.AppData
	Query  url.Values
	Errors validation.List
}

func Guidance(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		return tmpl(w, &guidanceData{
			App:   appData,
			Query: r.URL.Query(),
		})
	}
}
