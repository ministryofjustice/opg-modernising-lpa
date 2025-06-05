package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourAuthorisedSignatoryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *yourAuthorisedSignatoryForm
}

func YourAuthorisedSignatory(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourAuthorisedSignatoryData{
			App: appData,
			Form: &yourAuthorisedSignatoryForm{
				FirstNames: provided.AuthorisedSignatory.FirstNames,
				LastName:   provided.AuthorisedSignatory.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourAuthorisedSignatoryForm(r)
			data.Errors = data.Form.Validate()

			nameMatches := signatoryMatches(provided, data.Form.FirstNames, data.Form.LastName)
			redirectToWarning := false

			if !nameMatches.IsNone() && provided.AuthorisedSignatory.NameHasChanged(data.Form.FirstNames, data.Form.LastName) {
				redirectToWarning = true
			}

			if data.Errors.None() {
				if provided.AuthorisedSignatory.UID.IsZero() {
					provided.AuthorisedSignatory.UID = newUID()
				}

				provided.AuthorisedSignatory.FirstNames = data.Form.FirstNames
				provided.AuthorisedSignatory.LastName = data.Form.LastName

				if !provided.Tasks.ChooseYourSignatory.IsCompleted() {
					provided.Tasks.ChooseYourSignatory = task.StateInProgress
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {donor.PathYourIndependentWitness.Format(provided.LpaID)},
						"actor":       {actor.TypeAuthorisedSignatory.String()},
					})
				}

				return donor.PathYourIndependentWitness.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourAuthorisedSignatoryForm struct {
	FirstNames string
	LastName   string
}

func readYourAuthorisedSignatoryForm(r *http.Request) *yourAuthorisedSignatoryForm {
	return &yourAuthorisedSignatoryForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *yourAuthorisedSignatoryForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
