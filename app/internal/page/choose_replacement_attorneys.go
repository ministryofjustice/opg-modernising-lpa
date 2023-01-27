package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysData struct {
	App        AppData
	Errors     validation.List
	Form       *chooseAttorneysForm
	DobWarning string
}

func ChooseReplacementAttorneys(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		addAnother := r.FormValue("addAnother") == "1"
		ra, attorneyFound := lpa.GetReplacementAttorney(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.ReplacementAttorneys) > 0 && attorneyFound == false && addAnother == false {
			return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
		}

		data := &chooseReplacementAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: ra.FirstNames,
				LastName:   ra.LastName,
				Email:      ra.Email,
			},
		}

		if !ra.DateOfBirth.IsZero() {
			data.Form.Dob = readDate(ra.DateOfBirth)
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.Empty() && data.DobWarning == "" {
				if attorneyFound == false {
					ra = Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.DateOfBirth,
						ID:          randomString(8),
					}

					lpa.ReplacementAttorneys = append(lpa.ReplacementAttorneys, ra)
				} else {
					ra.FirstNames = data.Form.FirstNames
					ra.LastName = data.Form.LastName
					ra.Email = data.Form.Email
					ra.DateOfBirth = data.Form.DateOfBirth

					lpa.PutReplacementAttorney(ra)
				}

				if !attorneyFound {
					lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChooseReplacementAttorneysAddress, ra.ID)
				}

				return appData.Redirect(w, r, lpa, from)
			}
		}

		return tmpl(w, data)
	}
}
