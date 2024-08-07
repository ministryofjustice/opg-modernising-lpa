package donorpage

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type withdrawLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func WithdrawLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			provided.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return page.PathLpaWithdrawn.RedirectQuery(w, r, appData, url.Values{"uid": {provided.LpaUID}})
		}

		return tmpl(w, &withdrawLpaData{
			App:   appData,
			Donor: provided,
		})
	}
}
