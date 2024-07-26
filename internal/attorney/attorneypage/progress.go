package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type progressData struct {
	App             page.AppData
	Errors          validation.List
	Lpa             *lpastore.Lpa
	Signed          bool
	AttorneysSigned bool
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
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
