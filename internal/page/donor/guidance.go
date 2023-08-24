package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func Guidance(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &guidanceData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
