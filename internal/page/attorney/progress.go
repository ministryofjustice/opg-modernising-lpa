package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type progressData struct {
	App             page.AppData
	Errors          validation.List
	Donor           *actor.DonorProvidedDetails
	Signed          bool
	AttorneysSigned bool
}

func Progress(tmpl template.Template, attorneyStore AttorneyStore, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		donor, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		attorneys, err := attorneyStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		data := &progressData{
			App:             appData,
			Donor:           donor,
			Signed:          attorneyProvidedDetails.Signed(donor.SignedAt),
			AttorneysSigned: donor.AllAttorneysSigned(attorneys),
		}

		return tmpl(w, data)
	}
}
