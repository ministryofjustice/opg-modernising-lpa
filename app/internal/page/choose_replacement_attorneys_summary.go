package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App    AppData
	Errors validation.List
	Form   *chooseAttorneysSummaryForm
	Lpa    *Lpa
}

func ChooseReplacementAttorneysSummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: &chooseAttorneysSummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysSummaryForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.Empty() {
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

				return appData.Redirect(w, r, lpa, redirectUrl)
			}

		}

		return tmpl(w, data)
	}
}
