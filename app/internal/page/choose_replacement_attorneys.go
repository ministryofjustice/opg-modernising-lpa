package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseReplacementAttorneysData struct {
	App        AppData
	Errors     map[string]string
	Form       *chooseAttorneysForm
	DobWarning string
}

func ChooseReplacementAttorneys(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		addAnother := r.FormValue("addAnother") == "1"
		ra, attorneyFound := lpa.GetReplacementAttorney(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.ReplacementAttorneys) > 0 && attorneyFound == false && addAnother == false {
			appData.Lang.Redirect(w, r, chooseReplacementAttorneysSummaryPath, http.StatusFound)
			return nil
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

			if len(data.Errors) != 0 || data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if len(data.Errors) == 0 && data.DobWarning == "" {
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

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", chooseReplacementAttorneysAddressPath, ra.ID)
				}

				appData.Lang.Redirect(w, r, from, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}
