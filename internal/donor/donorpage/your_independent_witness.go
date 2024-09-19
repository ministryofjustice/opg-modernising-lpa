package donorpage

import (
	"net/http"
	"strings"

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
	App         appcontext.Data
	Errors      validation.List
	Form        *yourIndependentWitnessForm
	NameWarning *actor.SameNameWarning
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

			nameWarning := actor.NewSameNameWarning(
				actor.TypeIndependentWitness,
				independentWitnessMatches(provided, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if provided.IndependentWitness.UID.IsZero() {
					provided.IndependentWitness.UID = newUID()
				}

				provided.IndependentWitness.FirstNames = data.Form.FirstNames
				provided.IndependentWitness.LastName = data.Form.LastName

				if !provided.Tasks.ChooseYourSignatory.IsCompleted() {
					provided.Tasks.ChooseYourSignatory = task.StateInProgress
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathYourIndependentWitnessMobile.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourIndependentWitnessForm struct {
	FirstNames        string
	LastName          string
	IgnoreNameWarning string
}

func readYourIndependentWitnessForm(r *http.Request) *yourIndependentWitnessForm {
	return &yourIndependentWitnessForm{
		FirstNames:        page.PostFormString(r, "first-names"),
		LastName:          page.PostFormString(r, "last-name"),
		IgnoreNameWarning: page.PostFormString(r, "ignore-name-warning"),
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

func independentWitnessMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !person.Type.IsIndependentWitness() &&
			!person.Type.IsPersonToNotify() &&
			strings.EqualFold(person.FirstNames, firstNames) &&
			strings.EqualFold(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}
