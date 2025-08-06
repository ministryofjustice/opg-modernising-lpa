package donorpage

import (
	"net/http"
	"net/url"
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
	CanUpdateAddress bool
	FirstNamesMatch  bool
	LastNameMatch    bool
	DateOfBirthMatch bool
	AddressMatch     bool
	Form             *form.YesNoForm
}

func (d identityDetailsData) State() string {
	detailsMatched := d.FirstNamesMatch && d.LastNameMatch && d.DateOfBirthMatch

	if !d.Provided.CanChange() && !detailsMatched {
		return "cannotChange"
	} else if !detailsMatched && !d.Provided.ContinueWithMismatchedDetails {
		return "detailNotMatched"
	} else if !d.AddressMatch && d.CanUpdateAddress {
		return "addressNotMatched"
	} else {
		return "matched"
	}
}

func IdentityDetails(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &identityDetailsData{
			App:              appData,
			Form:             form.NewYesNoForm(form.YesNoUnknown),
			Provided:         provided,
			CanUpdateAddress: r.FormValue("canUpdateAddress") == "1",
			FirstNamesMatch:  strings.EqualFold(provided.Donor.FirstNames, provided.IdentityUserData.FirstNames),
			LastNameMatch:    strings.EqualFold(provided.Donor.LastName, provided.IdentityUserData.LastName),
			DateOfBirthMatch: provided.Donor.DateOfBirth == provided.IdentityUserData.DateOfBirth,
			AddressMatch:     provided.Donor.Address == provided.IdentityUserData.CurrentAddress,
		}

		if r.Method == http.MethodPost {
			errorLabel := "yesIfWouldLikeToUpdateDetails"
			if !provided.CanChange() {
				errorLabel = "yesToRevokeThisLpaAndMakeNew"
			} else if !data.AddressMatch {
				errorLabel = "anOptionForTheAddressInYourLpa"
			}

			f := form.ReadYesNoForm(r, errorLabel)
			data.Errors = f.Validate()

			if data.Errors.None() {
				var (
					redirect      donor.Path
					redirectQuery url.Values
				)

				switch data.State() {
				case "cannotChange":
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

				case "addressNotMatched":
					if f.YesNo.IsYes() {
						provided.Donor.Address = provided.IdentityUserData.CurrentAddress
						if !provided.ContinueWithMismatchedDetails {
							provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
						}
						provided.IdentityDetailsCausedCheck = true

						redirect = donor.PathIdentityDetailsUpdated
						redirectQuery = url.Values{"address": {"1"}}
					} else {
						if provided.ContinueWithMismatchedDetails {
							redirect = donor.PathRegisterWithCourtOfProtection
						} else {
							redirect = donor.PathTaskList
						}
					}

				case "detailNotMatched":
					if f.YesNo.IsYes() {
						provided.Donor.FirstNames = provided.IdentityUserData.FirstNames
						provided.Donor.LastName = provided.IdentityUserData.LastName
						provided.Donor.DateOfBirth = provided.IdentityUserData.DateOfBirth
						provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
						provided.IdentityDetailsCausedCheck = true

						if data.AddressMatch {
							redirect = donor.PathIdentityDetailsUpdated
						} else {
							redirect = donor.PathIdentityDetails
							redirectQuery = url.Values{"canUpdateAddress": {"1"}, "updated": {"1"}}
						}
					} else {
						provided.ContinueWithMismatchedDetails = true
						provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

						if data.AddressMatch {
							redirect = donor.PathRegisterWithCourtOfProtection
						} else {
							redirect = donor.PathIdentityDetails
							redirectQuery = url.Values{"canUpdateAddress": {"1"}, "notUpdated": {"1"}}
						}
					}
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if redirectQuery == nil {
					return redirect.Redirect(w, r, appData, provided)
				} else {
					return redirect.RedirectQuery(w, r, appData, provided, redirectQuery)
				}
			}
		}

		return tmpl(w, data)
	}
}
