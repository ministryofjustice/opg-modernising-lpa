package donor

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App         page.AppData
	Errors      validation.List
	Form        *chooseAttorneysForm
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func ChooseReplacementAttorneys(tmpl template.Template, donorStore DonorStore, uuidString func() string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := lpa.ReplacementAttorneys.Get(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && lpa.ReplacementAttorneys.Len() > 0 && !attorneyFound && !addAnother {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary.Format(lpa.ID))
		}

		data := &chooseReplacementAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
				Dob:        attorney.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeReplacementAttorney,
				replacementAttorneyMatches(lpa, attorney.ID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.DobWarning == "" && data.NameWarning == nil {
				if attorneyFound == false {
					attorney = actor.Attorney{ID: uuidString()}
				}

				attorney.FirstNames = data.Form.FirstNames
				attorney.LastName = data.Form.LastName
				attorney.Email = data.Form.Email
				attorney.DateOfBirth = data.Form.Dob

				lpa.ReplacementAttorneys.Put(attorney)

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, appData.Paths.ChooseReplacementAttorneysAddress.Format(lpa.ID)+"?id="+attorney.ID)
			}
		}

		return tmpl(w, data)
	}
}

func replacementAttorneyMatches(lpa *page.Lpa, id, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	if strings.EqualFold(lpa.Donor.FirstNames, firstNames) && strings.EqualFold(lpa.Donor.LastName, lastName) {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys.Attorneys() {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys() {
		if attorney.ID != id && strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(lpa.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(lpa.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	for _, person := range lpa.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, firstNames) && strings.EqualFold(person.LastName, lastName) {
			return actor.TypePersonToNotify
		}
	}

	return actor.TypeNone
}
