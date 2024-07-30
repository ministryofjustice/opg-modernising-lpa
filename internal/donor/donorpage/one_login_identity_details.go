package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App            page.AppData
	Errors         validation.List
	DonorProvided  *actor.DonorProvidedDetails
	DetailsMatch   bool
	DetailsUpdated bool
	Form           *form.YesNoForm
}

func OneLoginIdentityDetails(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &oneLoginIdentityDetailsData{
			App:            appData,
			Form:           form.NewYesNoForm(form.YesNoUnknown),
			DonorProvided:  donor,
			DetailsUpdated: r.FormValue("detailsUpdated") == "1",
			DetailsMatch: donor.Donor.FirstNames == donor.DonorIdentityUserData.FirstNames &&
				donor.Donor.LastName == donor.DonorIdentityUserData.LastName &&
				donor.Donor.DateOfBirth == donor.DonorIdentityUserData.DateOfBirth &&
				donor.Donor.Address.Postcode == donor.DonorIdentityUserData.CurrentAddress.Postcode,
		}

		if r.Method == http.MethodPost {
			if donor.DonorIdentityConfirmed() {
				return page.Paths.ReadYourLpa.Redirect(w, r, appData, donor)
			}

			f := form.ReadYesNoForm(r, "yesIfWouldLikeToUpdateDetails")
			data.Errors = f.Validate()

			if data.Errors.None() {
				if f.YesNo.IsYes() {
					donor.Donor.FirstNames = donor.DonorIdentityUserData.FirstNames
					donor.Donor.LastName = donor.DonorIdentityUserData.LastName
					donor.Donor.DateOfBirth = donor.DonorIdentityUserData.DateOfBirth
					donor.Donor.Address = donor.DonorIdentityUserData.CurrentAddress
					if err := donor.UpdateCheckedHash(); err != nil {
						return err
					}

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}

					return page.Paths.OneLoginIdentityDetails.RedirectQuery(w, r, appData, donor, url.Values{"detailsUpdated": {"1"}})
				} else {
					return page.Paths.WithdrawThisLpa.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}
