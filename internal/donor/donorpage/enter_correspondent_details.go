package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

func EnterCorrespondentDetails(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &enterCorrespondentDetailsData{
			App: appData,
			Form: &enterCorrespondentDetailsForm{
				FirstNames:   donor.Correspondent.FirstNames,
				LastName:     donor.Correspondent.LastName,
				Email:        donor.Correspondent.Email,
				Organisation: donor.Correspondent.Organisation,
				Telephone:    donor.Correspondent.Telephone,
				WantAddress:  form.NewYesNoForm(donor.Correspondent.WantAddress),
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterCorrespondentDetailsForm(r, donor.Donor)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Correspondent.FirstNames = data.Form.FirstNames
				donor.Correspondent.LastName = data.Form.LastName
				donor.Correspondent.Email = data.Form.Email
				donor.Correspondent.Organisation = data.Form.Organisation
				donor.Correspondent.Telephone = data.Form.Telephone
				donor.Correspondent.WantAddress = data.Form.WantAddress.YesNo

				var redirect page.LpaPath
				if donor.Correspondent.WantAddress.IsNo() {
					donor.Correspondent.Address = place.Address{}
					donor.Tasks.AddCorrespondent = task.StateCompleted
					redirect = page.Paths.TaskList
				} else {
					if !donor.Tasks.AddCorrespondent.Completed() && donor.Correspondent.Address.Line1 == "" {
						donor.Tasks.AddCorrespondent = task.StateInProgress
					}
					redirect = page.Paths.EnterCorrespondentAddress
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, donor)
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
	Telephone       string
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
		Telephone:       page.PostFormString(r, "telephone"),
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

	errors.String("telephone", "phoneNumber", f.Telephone,
		validation.Telephone())

	return errors.Append(f.WantAddress.Validate())
}