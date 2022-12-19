package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseReplacementAttorneysSummaryData struct {
	App    AppData
	Errors map[string]string
	Form   chooseAttorneysSummaryForm
	Lpa    *Lpa
}

func ChooseReplacementAttorneysSummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: chooseAttorneysSummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = chooseAttorneysSummaryForm{
				AddAttorney: postFormString(r, "add-attorney"),
			}

			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				var redirectUrl string

				if len(lpa.ReplacementAttorneys) > 1 && len(lpa.Attorneys) > 1 && lpa.HowAttorneysMakeDecisions == Jointly {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysMakeDecisions
				} else if len(lpa.Attorneys) > 1 && lpa.HowAttorneysMakeDecisions == JointlyAndSeverally {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysStepIn
				} else {
					redirectUrl = appData.Paths.WhenCanTheLpaBeUsed
				}

				if data.Form.AddAttorney == "yes" {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseReplacementAttorneys)
				}

				return appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
			}

		}

		return tmpl(w, data)
	}
}
