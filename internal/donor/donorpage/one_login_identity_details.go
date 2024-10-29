package donorpage

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type oneLoginIdentityDetailsData struct {
	App              appcontext.Data
	Errors           validation.List
	Provided         *donordata.Provided
	FirstNamesMatch  bool
	LastNameMatch    bool
	DateOfBirthMatch bool
	AddressMatch     bool
	DetailsUpdated   bool
	Form             *form.YesNoForm
}

func (d oneLoginIdentityDetailsData) DetailsMatch() bool {
	return d.FirstNamesMatch && d.LastNameMatch && d.DateOfBirthMatch && d.AddressMatch
}

func OneLoginIdentityDetails(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &oneLoginIdentityDetailsData{
			App:              appData,
			Form:             form.NewYesNoForm(form.YesNoUnknown),
			Provided:         provided,
			DetailsUpdated:   r.FormValue("detailsUpdated") == "1",
			FirstNamesMatch:  strings.EqualFold(provided.Donor.FirstNames, provided.IdentityUserData.FirstNames),
			LastNameMatch:    strings.EqualFold(provided.Donor.LastName, provided.IdentityUserData.LastName),
			DateOfBirthMatch: provided.Donor.DateOfBirth == provided.IdentityUserData.DateOfBirth,
			AddressMatch:     provided.Donor.Address.Postcode == provided.IdentityUserData.CurrentAddress.Postcode,
		}

		if r.Method == http.MethodPost {
			if provided.DonorIdentityConfirmed() {
				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}

			f := form.ReadYesNoForm(r, "yesIfWouldLikeToUpdateDetails")
			data.Errors = f.Validate()

			if data.Errors.None() {
				if f.YesNo.IsYes() {
					provided.Donor.FirstNames = provided.IdentityUserData.FirstNames
					provided.Donor.LastName = provided.IdentityUserData.LastName
					provided.Donor.DateOfBirth = provided.IdentityUserData.DateOfBirth
					provided.Donor.Address = provided.IdentityUserData.CurrentAddress
					provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
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
