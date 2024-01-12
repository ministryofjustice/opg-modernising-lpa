package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type MakeANewLPAData struct {
	App         page.AppData
	Errors      validation.List
	FullName    string
	DateOfBirth date.Date
	Address     place.Address
}

func MakeANewLPA(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		previouslyProvidedDetails, err := donorStore.Latest(r.Context())
		if err != nil {
			return err
		}

		data := MakeANewLPAData{
			App:         appData,
			FullName:    previouslyProvidedDetails.Donor.FullName(),
			DateOfBirth: previouslyProvidedDetails.Donor.DateOfBirth,
			Address:     previouslyProvidedDetails.Donor.Address,
		}

		if previouslyProvidedDetails.Donor.FullName() != donor.Donor.FullName() && donor.Donor.FirstNames != "" {
			data.FullName = donor.Donor.FullName()
		}

		if previouslyProvidedDetails.Donor.DateOfBirth != donor.Donor.DateOfBirth && !donor.Donor.DateOfBirth.IsZero() {
			data.DateOfBirth = donor.Donor.DateOfBirth
		}

		if previouslyProvidedDetails.Donor.Address != donor.Donor.Address && donor.Donor.Address.String() != "" {
			data.Address = donor.Donor.Address
		}

		return tmpl(w, data)
	}
}
