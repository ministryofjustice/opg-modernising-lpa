package page

import (
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
			appData.Lang.Redirect(w, r, chooseAttorneysSummaryPath, http.StatusFound)
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

		return tmpl(w, data)
	}
}
