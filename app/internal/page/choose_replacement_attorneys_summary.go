package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseReplacementAttorneysSummaryData struct {
	App                            AppData
	ReplacementAttorneyAddressPath string
	ReplacementAttorneyDetailsPath string
	Errors                         map[string]string
	Form                           chooseAttorneysSummaryForm
	Lpa                            *Lpa
	RemoveReplacementAttorneyPath  string
}

func ChooseReplacementAttorneysSummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:                            appData,
			Lpa:                            lpa,
			ReplacementAttorneyDetailsPath: chooseReplacementAttorneysPath,
			ReplacementAttorneyAddressPath: chooseReplacementAttorneysAddressPath,
			Form:                           chooseAttorneysSummaryForm{},
			RemoveReplacementAttorneyPath:  removeReplacementAttorneyPath,
		}

		if r.Method == http.MethodPost {
			data.Form = chooseAttorneysSummaryForm{
				AddAttorney: postFormString(r, "add-attorney"),
			}

			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				var redirectUrl string

				if len(lpa.ReplacementAttorneys) > 1 && len(lpa.Attorneys) > 1 && lpa.DecisionsType == "jointly" {
					redirectUrl = howShouldReplacementAttorneysMakeDecisionsPath
				} else if len(lpa.Attorneys) > 1 && lpa.DecisionsType == "jointly-and-severally" {
					redirectUrl = howShouldReplacementAttorneysStepInPath
				} else {
					redirectUrl = taskListPath
				}

				if data.Form.AddAttorney == "yes" {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", data.ReplacementAttorneyDetailsPath)
				}

				appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
				return nil
			}

		}

		return tmpl(w, data)
	}
}
