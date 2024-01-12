package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type MakeANewLPAData struct {
	App                page.AppData
	Errors             validation.List
	LatestDonorDetails *actor.DonorProvidedDetails
}

func MakeANewLPA(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		latestDonorDetails, err := donorStore.Latest(r.Context())
		if err != nil {
			return err
		}

		if latestDonorDetails.Donor.FullName() != donor.Donor.FullName() && donor.Donor.FirstNames != "" {
			latestDonorDetails.Donor.FirstNames = donor.Donor.FirstNames
			latestDonorDetails.Donor.LastName = donor.Donor.LastName
		}

		if latestDonorDetails.Donor.DateOfBirth != donor.Donor.DateOfBirth && !donor.Donor.DateOfBirth.IsZero() {
			latestDonorDetails.Donor.DateOfBirth = donor.Donor.DateOfBirth
		}

		if latestDonorDetails.Donor.Address != donor.Donor.Address && donor.Donor.Address.String() != "" {
			latestDonorDetails.Donor.Address = donor.Donor.Address
		}

		return tmpl(w, MakeANewLPAData{
			App:                appData,
			LatestDonorDetails: latestDonorDetails,
		})
	}
}
