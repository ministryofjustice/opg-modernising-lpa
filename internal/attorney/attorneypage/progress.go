package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type progressData struct {
	App             appcontext.Data
	Errors          validation.List
	Lpa             *lpadata.Lpa
	Signed          bool
	AttorneysSigned bool
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &progressData{
			App:             appData,
			Lpa:             lpa,
			Signed:          attorneyProvidedDetails.Signed(),
			AttorneysSigned: lpa.AllAttorneysSigned(),
		}

		return tmpl(w, data)
	}
}
