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

type yourIndependentWitnessData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *yourIndependentWitnessForm
}

func YourIndependentWitness(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourIndependentWitnessData{
			App: appData,
			Form: &yourIndependentWitnessForm{
				FirstNames: provided.IndependentWitness.FirstNames,
				LastName:   provided.IndependentWitness.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			nameMatches := independentWitnessMatches(provided, data.Form.FirstNames, data.Form.LastName)
			redirectToWarning := false

			if !nameMatches.IsNone() && provided.IndependentWitness.NameHasChanged(data.Form.FirstNames, data.Form.LastName) {
				redirectToWarning = true
			}

			if data.Errors.None() {
				if provided.IndependentWitness.UID.IsZero() {
					provided.IndependentWitness.UID = newUID()
				}

				provided.IndependentWitness.FirstNames = data.Form.FirstNames
				provided.IndependentWitness.LastName = data.Form.LastName

				if !provided.Tasks.ChooseYourSignatory.IsCompleted() {
					provided.Tasks.ChooseYourSignatory = task.StateInProgress
				}

				// Allow changing details for independent witness on the page they
				// witness, without certificate provider having to be notified.
				if !provided.SignedAt.IsZero() {
					provided.UpdateCheckedHash()
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {donor.PathYourIndependentWitnessMobile.Format(provided.LpaID)},
						"actor":       {actor.TypeIndependentWitness.String()},
					})
				}

				return donor.PathYourIndependentWitnessMobile.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourIndependentWitnessForm struct {
	FirstNames string
	LastName   string
}

func readYourIndependentWitnessForm(r *http.Request) *yourIndependentWitnessForm {
	return &yourIndependentWitnessForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *yourIndependentWitnessForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
