package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterCorrespondentDetailsData struct {
	App               appcontext.Data
	Errors            validation.List
	Form              *enterCorrespondentDetailsForm
	CompletedAllTasks bool
}

func EnterCorrespondentDetails(tmpl template.Template, service CorrespondentService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &enterCorrespondentDetailsData{
			App:               appData,
			Form:              newEnterCorrespondentDetailsForm(appData.Localizer),
			CompletedAllTasks: provided.CompletedAllTasks(),
		}

		data.Form.FirstNames.SetInput(provided.Correspondent.FirstNames)
		data.Form.LastName.SetInput(provided.Correspondent.LastName)
		data.Form.Email.SetInput(provided.Correspondent.Email)
		data.Form.Organisation.SetInput(provided.Correspondent.Organisation)
		data.Form.Phone.SetInput(provided.Correspondent.Phone)
		data.Form.WantAddress.SetInput(provided.Correspondent.WantAddress)

		if r.Method == http.MethodPost {
			ok := data.Form.Parse(r, appData.Localizer, provided.Donor)

			nameMatchesDonor := correspondentNameMatchesDonor(provided, data.Form.FirstNames.Value, data.Form.LastName.Value)
			redirectToWarning := false

			if provided.Correspondent.NameHasChanged(data.Form.FirstNames.Value, data.Form.LastName.Value) && nameMatchesDonor {
				redirectToWarning = true
			}

			if ok {
				provided.Correspondent.FirstNames = data.Form.FirstNames.Value
				provided.Correspondent.LastName = data.Form.LastName.Value
				provided.Correspondent.Email = data.Form.Email.Value
				provided.Correspondent.Organisation = data.Form.Organisation.Value
				provided.Correspondent.Phone = data.Form.Phone.Value
				wantedAddress := provided.Correspondent.WantAddress
				provided.Correspondent.WantAddress = data.Form.WantAddress.Value

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
	newforms.Form
	FirstNames   *newforms.String
	LastName     *newforms.String
	Email        *newforms.String
	Organisation *newforms.String
	Phone        *newforms.String
	WantAddress  *newforms.YesNo

	DonorEmailMatch bool
	DonorFullName   string
}

func newEnterCorrespondentDetailsForm(l Localizer) *enterCorrespondentDetailsForm {
	return &enterCorrespondentDetailsForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
		Email: newforms.NewString("email", l.T("email")).
			NotEmpty().
			Email(),
		Organisation: newforms.NewString("organisation", ""),
		Phone: newforms.NewString("phone", l.T("phoneNumber")).
			Phone(),
		WantAddress: newforms.NewYesNo(l.T("yesToAddAnAddress")),
	}
}

func (f *enterCorrespondentDetailsForm) Parse(r *http.Request, l Localizer, donor donordata.Donor) bool {
	ok := f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
		f.Email,
		f.Organisation,
		f.Phone,
		f.WantAddress,
	)

	if f.Email.Value == donor.Email {
		f.Email.Error = newforms.FormattedError{
			Key:  "youProvidedThisEmailForDonorError",
			Data: map[string]any{"DonorFullName": f.DonorFullName},
		}
		f.Errors = append(f.Errors, f.Email.Field)
		ok = false
	}

	return ok
}
