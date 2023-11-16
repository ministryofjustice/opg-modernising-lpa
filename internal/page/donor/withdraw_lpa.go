package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type withdrawLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WithdrawLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			lpa.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, page.Paths.LpaWithdrawn.Format()+"?uid="+lpa.UID)
		}

		return tmpl(w, &withdrawLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
