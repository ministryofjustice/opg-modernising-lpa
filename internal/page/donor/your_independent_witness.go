package donor

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourIndependentWitnessData struct {
	App         page.AppData
	Errors      validation.List
	Form        *yourIndependentWitnessForm
	NameWarning *actor.SameNameWarning
}

func YourIndependentWitness(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		data := &yourIndependentWitnessData{
			App: appData,
			Form: &yourIndependentWitnessForm{
				FirstNames: lpa.IndependentWitness.FirstNames,
				LastName:   lpa.IndependentWitness.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeIndependentWitness,
				independentWitnessMatches(lpa, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if !data.Errors.Any() && data.NameWarning == nil {
				lpa.IndependentWitness.FirstNames = data.Form.FirstNames
				lpa.IndependentWitness.LastName = data.Form.LastName

				if !lpa.Tasks.ChooseYourSignatory.Completed() {
					lpa.Tasks.ChooseYourSignatory = actor.TaskInProgress
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return page.Paths.YourIndependentWitnessMobile.Redirect(w, r, appData, lpa)
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

func independentWitnessMatches(lpa *actor.DonorProvidedDetails, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	if strings.EqualFold(lpa.Donor.FirstNames, firstNames) && strings.EqualFold(lpa.Donor.LastName, lastName) {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(lpa.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(lpa.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	if strings.EqualFold(lpa.AuthorisedSignatory.FirstNames, firstNames) && strings.EqualFold(lpa.AuthorisedSignatory.LastName, lastName) {
		return actor.TypeAuthorisedSignatory
	}

	return actor.TypeNone
}
