package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterCorrespondentDetailsData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterCorrespondentDetailsForm
}

func EnterCorrespondentDetails(tmpl template.Template, service CorrespondentService) Handler {
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

			nameMatchesDonor := correspondentNameMatchesDonor(provided, data.Form.FirstNames, data.Form.LastName)
			redirectToWarning := false

			if provided.Correspondent.NameHasChanged(data.Form.FirstNames, data.Form.LastName) && nameMatchesDonor {
				redirectToWarning = true
			}

			if data.Errors.None() {
				provided.Correspondent.FirstNames = data.Form.FirstNames
				provided.Correspondent.LastName = data.Form.LastName
				provided.Correspondent.Email = data.Form.Email
				provided.Correspondent.Organisation = data.Form.Organisation
				provided.Correspondent.Phone = data.Form.Phone
				wantedAddress := provided.Correspondent.WantAddress
				provided.Correspondent.WantAddress = data.Form.WantAddress.YesNo

				var redirect donor.Path
				if provided.Correspondent.WantAddress.IsNo() {
					redirect = donor.PathCorrespondentSummary
				} else {
					redirect = donor.PathEnterCorrespondentAddress
				}

				if err := service.Put(r.Context(), provided); err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {redirect.Format(provided.LpaID)},
						"actor":       {actor.TypeCorrespondent.String()},
					})
				}

				if !wantedAddress.IsYes() && provided.Correspondent.WantAddress.IsYes() {
					from := r.FormValue("from")
					delete(r.Form, "from")
					return redirect.RedirectQuery(w, r, appData, provided, url.Values{
						"from": {from},
					})
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
