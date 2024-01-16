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

type makeANewLPAData struct {
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

		data := makeANewLPAData{
			App:         appData,
			FullName:    previouslyProvidedDetails.Donor.FullName(),
			DateOfBirth: previouslyProvidedDetails.Donor.DateOfBirth,
			Address:     previouslyProvidedDetails.Donor.Address,
		}

		if data.FullName != donor.Donor.FullName() && donor.Donor.FirstNames != "" {
			data.FullName = donor.Donor.FullName()
		} else {
			donor.Donor.FirstNames = previouslyProvidedDetails.Donor.FirstNames
			donor.Donor.LastName = previouslyProvidedDetails.Donor.LastName
			donor.Donor.OtherNames = previouslyProvidedDetails.Donor.OtherNames
		}

		if data.DateOfBirth != donor.Donor.DateOfBirth && !donor.Donor.DateOfBirth.IsZero() {
			data.DateOfBirth = donor.Donor.DateOfBirth
		} else {
			donor.Donor.DateOfBirth = previouslyProvidedDetails.Donor.DateOfBirth
		}

		if data.Address != donor.Donor.Address && donor.Donor.Address.Line1 != "" {
			data.Address = donor.Donor.Address
		} else {
			donor.Donor.Address = previouslyProvidedDetails.Donor.Address
		}

		if r.Method == http.MethodPost {
			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			return page.Paths.YourDetails.Redirect(w, r, appData, donor)
		}

		return tmpl(w, data)
	}
}
