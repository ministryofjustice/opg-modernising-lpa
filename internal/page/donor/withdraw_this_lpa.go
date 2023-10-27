package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type withdrawThisLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WithdrawThisLpa(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			// withdraw
			return appData.Redirect(w, r, nil, page.Paths.LpaWithdrawn.Format()+"?uid="+lpa.UID)
		}

		return tmpl(w, &withdrawThisLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
