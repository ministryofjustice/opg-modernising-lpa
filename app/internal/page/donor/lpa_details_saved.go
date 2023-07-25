package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type LpaDetailsSavedData struct {
	App          page.AppData
	Lpa          *page.Lpa
	IsFirstCheck bool
	Errors       validation.List
}

func LpaDetailsSaved(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		return tmpl(w, LpaDetailsSavedData{
			App:          appData,
			IsFirstCheck: r.URL.Query().Has("firstCheck"),
			Lpa:          lpa,
		})
	}
}
