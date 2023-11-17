package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type LpaDetailsSavedData struct {
	App          page.AppData
	Lpa          *actor.DonorProvidedDetails
	IsFirstCheck bool
	Errors       validation.List
}

func LpaDetailsSaved(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		return tmpl(w, LpaDetailsSavedData{
			App:          appData,
			IsFirstCheck: r.URL.Query().Has("firstCheck"),
			Lpa:          lpa,
		})
	}
}
