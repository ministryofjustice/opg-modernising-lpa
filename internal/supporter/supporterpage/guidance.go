package supporterpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App          appcontext.Data
	Query        url.Values
	Errors       validation.List
	Organisation *actor.Organisation
}

func Guidance(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		return tmpl(w, &guidanceData{
			App:          appData,
			Query:        r.URL.Query(),
			Organisation: organisation,
		})
	}
}
