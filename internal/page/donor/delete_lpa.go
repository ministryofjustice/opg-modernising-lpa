package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type deleteLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func DeleteLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			if err := donorStore.Delete(r.Context()); err != nil {
				return err
			}

			return appData.Redirect(w, r, page.Paths.LpaDeleted.Format()+"?uid="+lpa.UID)
		}

		return tmpl(w, &deleteLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
