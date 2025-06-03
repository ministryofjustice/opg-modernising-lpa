package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityDetailsData struct {
	App              appcontext.Data
	Errors           validation.List
	Provided         *donordata.Provided
	FirstNamesMatch  bool
	LastNameMatch    bool
	DateOfBirthMatch bool
	AddressMatch     bool
	Form             *form.YesNoForm
}

func (d identityDetailsData) DetailsMatch() bool {
	return d.FirstNamesMatch && d.LastNameMatch && d.DateOfBirthMatch && d.AddressMatch
}

func IdentityDetails(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &identityDetailsData{
			App:              appData,
			Form:             form.NewYesNoForm(form.YesNoUnknown),
			Provided:         provided,
			FirstNamesMatch:  strings.EqualFold(provided.Donor.FirstNames, provided.IdentityUserData.FirstNames),
			LastNameMatch:    strings.EqualFold(provided.Donor.LastName, provided.IdentityUserData.LastName),
			DateOfBirthMatch: provided.Donor.DateOfBirth == provided.IdentityUserData.DateOfBirth,
			AddressMatch:     provided.Donor.Address.Postcode == provided.IdentityUserData.CurrentAddress.Postcode,
		}

		if r.Method == http.MethodPost {
			errorLabel := "yesIfWouldLikeToUpdateDetails"
			if !provided.CanChange() {
				errorLabel = "yesToRevokeThisLpaAndMakeNew"
			}

			f := form.ReadYesNoForm(r, errorLabel)
			data.Errors = f.Validate()

			if data.Errors.None() {
				var redirect donor.Path

				if provided.CanChange() {
					if f.YesNo.IsYes() {
						provided.Donor.FirstNames = provided.IdentityUserData.FirstNames
						provided.Donor.LastName = provided.IdentityUserData.LastName
						provided.Donor.DateOfBirth = provided.IdentityUserData.DateOfBirth
						provided.Donor.Address = provided.IdentityUserData.CurrentAddress
						provided.Tasks.CheckYourLpa = task.StateInProgress
						provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
						provided.IdentityDetailsCausedCheck = true

						redirect = donor.PathIdentityDetailsUpdated
					} else {
						provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

						redirect = donor.PathRegisterWithCourtOfProtection
					}
				} else {
					if f.YesNo.IsYes() {
						return donor.PathWithdrawThisLpa.Redirect(w, r, appData, provided)
					} else {
						provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending
						provided.RegisteringWithCourtOfProtection = true

						redirect = donor.PathWhatHappensNextRegisteringWithCourtOfProtection

						if err := eventClient.SendRegisterWithCourtOfProtection(r.Context(), event.RegisterWithCourtOfProtection{
							UID: provided.LpaUID,
						}); err != nil {
							return err
						}
					}
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
