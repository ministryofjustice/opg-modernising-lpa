package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App         page.AppData
	Errors      validation.List
	Form        *chooseAttorneysForm
	DobWarning  string
	NameWarning *actor.SameNameWarning
}

func ChooseReplacementAttorneys(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		addAnother := r.FormValue("addAnother") == "1"
		attorney, attorneyFound := lpa.ReplacementAttorneys.Get(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.ReplacementAttorneys) > 0 && attorneyFound == false && addAnother == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
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
					attorney = actor.Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.Dob,
						ID:          randomString(8),
					}

					lpa.ReplacementAttorneys = append(lpa.ReplacementAttorneys, attorney)
				} else {
					attorney.FirstNames = data.Form.FirstNames
					attorney.LastName = data.Form.LastName
					attorney.Email = data.Form.Email
					attorney.DateOfBirth = data.Form.Dob

					lpa.ReplacementAttorneys.Put(attorney)
				}

				if !attorneyFound {
					lpa.Tasks.ChooseReplacementAttorneys = page.TaskInProgress
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChooseReplacementAttorneysAddress, attorney.ID)
				}

				return appData.Redirect(w, r, lpa, from)
			}
		}

		return tmpl(w, data)
	}
}

func replacementAttorneyMatches(lpa *page.Lpa, id, firstNames, lastName string) actor.Type {
	if lpa.You.FirstNames == firstNames && lpa.You.LastName == lastName {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys {
		if attorney.ID != id && attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeReplacementAttorney
		}
	}

	if lpa.CertificateProvider.FirstNames == firstNames && lpa.CertificateProvider.LastName == lastName {
		return actor.TypeCertificateProvider
	}

	for _, person := range lpa.PeopleToNotify {
		if person.FirstNames == firstNames && person.LastName == lastName {
			return actor.TypePersonToNotify
		}
	}

	return actor.TypeNone
}
