package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterCorrespondentDetailsData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *enterCorrespondentDetailsForm
	NameWarning *actor.SameNameWarning
}

func EnterCorrespondentDetails(tmpl template.Template, donorStore DonorStore, eventClient EventClient, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &enterCorrespondentDetailsData{
			App: appData,
			Form: &enterCorrespondentDetailsForm{
				FirstNames:   provided.Correspondent.FirstNames,
				LastName:     provided.Correspondent.LastName,
				Email:        provided.Correspondent.Email,
				Organisation: provided.Correspondent.Organisation,
				Phone:        provided.Correspondent.Phone,
				WantAddress:  form.NewYesNoForm(provided.Correspondent.WantAddress),
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterCorrespondentDetailsForm(r, provided.Donor)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.Correspondent.UID.IsZero() {
					provided.Correspondent.UID = newUID()
				}

				provided.Correspondent.FirstNames = data.Form.FirstNames
				provided.Correspondent.LastName = data.Form.LastName
				provided.Correspondent.Email = data.Form.Email
				provided.Correspondent.Organisation = data.Form.Organisation
				provided.Correspondent.Phone = data.Form.Phone
				provided.Correspondent.WantAddress = data.Form.WantAddress.YesNo

				var redirect donor.Path
				if provided.Correspondent.WantAddress.IsNo() {
					provided.Correspondent.Address = place.Address{}
					provided.Tasks.AddCorrespondent = task.StateCompleted

					if err := eventClient.SendCorrespondentUpdated(r.Context(), event.CorrespondentUpdated{
						UID:        provided.LpaUID,
						ActorUID:   provided.Correspondent.UID,
						FirstNames: provided.Correspondent.FirstNames,
						LastName:   provided.Correspondent.LastName,
						Email:      provided.Correspondent.Email,
						Phone:      provided.Correspondent.Phone,
					}); err != nil {
						return err
					}

					redirect = donor.PathTaskList
				} else {
					if !provided.Tasks.AddCorrespondent.IsCompleted() && provided.Correspondent.Address.Line1 == "" {
						provided.Tasks.AddCorrespondent = task.StateInProgress
					}
					redirect = donor.PathEnterCorrespondentAddress
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

type enterCorrespondentDetailsForm struct {
	FirstNames      string
	LastName        string
	Email           string
	Organisation    string
	Phone           string
	WantAddress     *form.YesNoForm
	DonorEmailMatch bool
	DonorFullName   string
}

func readEnterCorrespondentDetailsForm(r *http.Request, donor donordata.Donor) *enterCorrespondentDetailsForm {
	email := page.PostFormString(r, "email")

	return &enterCorrespondentDetailsForm{
		FirstNames:      page.PostFormString(r, "first-names"),
		LastName:        page.PostFormString(r, "last-name"),
		Email:           page.PostFormString(r, "email"),
		Organisation:    page.PostFormString(r, "organisation"),
		Phone:           page.PostFormString(r, "phone"),
		WantAddress:     form.ReadYesNoForm(r, "yesToAddAnAddress"),
		DonorEmailMatch: email == donor.Email,
		DonorFullName:   donor.FullName(),
	}
}

func (f *enterCorrespondentDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	if f.DonorEmailMatch {
		errors.Add("email", validation.CustomFormattedError{
			Label: "youProvidedThisEmailForDonorError",
			Data:  map[string]any{"DonorFullName": f.DonorFullName},
		})
	}

	errors.String("phone", "phoneNumber", f.Phone,
		validation.Phone())

	return errors.Append(f.WantAddress.Validate())
}
