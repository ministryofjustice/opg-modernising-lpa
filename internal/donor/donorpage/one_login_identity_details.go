package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App            appcontext.Data
	Errors         validation.List
	DonorProvided  *donordata.Provided
	DetailsMatch   bool
	DetailsUpdated bool
	Form           *form.YesNoForm
}

func OneLoginIdentityDetails(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &oneLoginIdentityDetailsData{
			App:            appData,
			Form:           form.NewYesNoForm(form.YesNoUnknown),
			DonorProvided:  provided,
			DetailsUpdated: r.FormValue("detailsUpdated") == "1",
			DetailsMatch: provided.Donor.FirstNames == provided.DonorIdentityUserData.FirstNames &&
				provided.Donor.LastName == provided.DonorIdentityUserData.LastName &&
				provided.Donor.DateOfBirth == provided.DonorIdentityUserData.DateOfBirth &&
				provided.Donor.Address.Postcode == provided.DonorIdentityUserData.CurrentAddress.Postcode,
		}

		if r.Method == http.MethodPost {
			if provided.DonorIdentityConfirmed() {
				return donor.PathReadYourLpa.Redirect(w, r, appData, provided)
			}

			f := form.ReadYesNoForm(r, "yesIfWouldLikeToUpdateDetails")
			data.Errors = f.Validate()

			if data.Errors.None() {
				if f.YesNo.IsYes() {
					provided.Donor.FirstNames = provided.DonorIdentityUserData.FirstNames
					provided.Donor.LastName = provided.DonorIdentityUserData.LastName
					provided.Donor.DateOfBirth = provided.DonorIdentityUserData.DateOfBirth
					provided.Donor.Address = provided.DonorIdentityUserData.CurrentAddress
					if err := provided.UpdateCheckedHash(); err != nil {
						return err
					}

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}

					return donor.PathOneLoginIdentityDetails.RedirectQuery(w, r, appData, provided, url.Values{"detailsUpdated": {"1"}})
				} else {
					return donor.PathWithdrawThisLpa.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
