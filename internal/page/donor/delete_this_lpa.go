package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteThisLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func DeleteThisLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			if err := donorStore.Delete(r.Context()); err != nil {
				return err
			}

			return appData.Redirect(w, r, nil, page.Paths.LpaDeleted.Format()+"?uid="+lpa.UID)
		}

		return tmpl(w, &deleteThisLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
