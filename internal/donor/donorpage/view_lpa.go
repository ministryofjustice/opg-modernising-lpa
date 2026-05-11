package donorpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type viewLpaData struct {
	App appcontext.Data
	Lpa *lpadata.Lpa
}

func ViewLpa(tmpl template.Template, lpaStoreClient LpaStoreClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreClient.LpaWithImages(r.Context(), donor.LpaUID)
		if err != nil {
			if !errors.Is(err, lpastore.ErrNotFound) {
				return err
			}

			lpa = lpastore.LpaFromDonorProvided(donor)
		}

		return tmpl(w, &viewLpaData{
			App: appData,
			Lpa: lpa,
		})
	}
}
