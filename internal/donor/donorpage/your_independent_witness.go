package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &yourIndependentWitnessData{
			App: appData,
			Form: &yourIndependentWitnessForm{
				FirstNames: donor.IndependentWitness.FirstNames,
				LastName:   donor.IndependentWitness.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourIndependentWitnessForm(r)
			data.Errors = data.Form.Validate()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeIndependentWitness,
				independentWitnessMatches(donor, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				donor.IndependentWitness.FirstNames = data.Form.FirstNames
				donor.IndependentWitness.LastName = data.Form.LastName

				if !donor.Tasks.ChooseYourSignatory.Completed() {
					donor.Tasks.ChooseYourSignatory = actor.TaskInProgress
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.YourIndependentWitnessMobile.Redirect(w, r, appData, donor)
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

func independentWitnessMatches(donor *donordata.DonorProvidedDetails, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	if strings.EqualFold(donor.Donor.FirstNames, firstNames) && strings.EqualFold(donor.Donor.LastName, lastName) {
		return actor.TypeDonor
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(donor.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(donor.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, firstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, lastName) {
		return actor.TypeAuthorisedSignatory
	}

	return actor.TypeNone
}
