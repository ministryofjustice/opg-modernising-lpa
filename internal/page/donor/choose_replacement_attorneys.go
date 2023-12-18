package donor

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App         page.AppData
	Errors      validation.List
	Donor       *actor.DonorProvidedDetails
	Form        *chooseAttorneysForm
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func ChooseReplacementAttorneys(tmpl template.Template, donorStore DonorStore, uuidString func() string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := donor.ReplacementAttorneys.Get(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && donor.ReplacementAttorneys.Len() > 0 && !attorneyFound && !addAnother {
			return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, donor)
		}

		data := &chooseReplacementAttorneysData{
			App:   appData,
			Donor: donor,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
		}

		if !attorney.DateOfBirth.IsZero() {
			data.DobWarning = data.Form.DobWarning()
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeReplacementAttorney,
				replacementAttorneyMatches(donor, attorney.ID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || dobWarning != "" {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = actor.Attorney{ID: uuidString()}
				}

				attorney.FirstNames = data.Form.FirstNames
				attorney.LastName = data.Form.LastName
				attorney.Email = data.Form.Email
				attorney.DateOfBirth = data.Form.Dob

				donor.ReplacementAttorneys.Put(attorney)

				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return appData.Paths.ChooseReplacementAttorneysAddress.RedirectQuery(w, r, appData, donor, url.Values{"id": {attorney.ID}})
			}
		}

		return tmpl(w, data)
	}
}

func replacementAttorneyMatches(donor *actor.DonorProvidedDetails, id, firstNames, lastName string) actor.Type {
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
		if attorney.ID != id && strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(donor.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(donor.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	for _, person := range donor.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, firstNames) && strings.EqualFold(person.LastName, lastName) {
			return actor.TypePersonToNotify
		}
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, firstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, lastName) {
		return actor.TypeAuthorisedSignatory
	}

	if strings.EqualFold(donor.IndependentWitness.FirstNames, firstNames) && strings.EqualFold(donor.IndependentWitness.LastName, lastName) {
		return actor.TypeIndependentWitness
	}

	return actor.TypeNone
}
