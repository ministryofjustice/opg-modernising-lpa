package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type withdrawThisLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
}

func WithdrawThisLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			lpa.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			return appData.Redirect(w, r, nil, page.Paths.LpaWithdrawn.Format()+"?uid="+lpa.UID)
		}

		return tmpl(w, &withdrawThisLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
