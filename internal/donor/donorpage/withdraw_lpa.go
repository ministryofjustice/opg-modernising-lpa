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

func WithdrawLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time, lpaStoreClient LpaStoreClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if r.Method == http.MethodPost {
			donor.WithdrawnAt = now()
			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			if err := lpaStoreClient.SendDonorWithdrawLPA(r.Context(), donor.LpaUID); err != nil {
				return err
			}

			return page.Paths.LpaWithdrawn.RedirectQuery(w, r, appData, url.Values{"uid": {donor.LpaUID}})
		}

		return tmpl(w, &withdrawLpaData{
			App:   appData,
			Donor: donor,
		})
	}
}
