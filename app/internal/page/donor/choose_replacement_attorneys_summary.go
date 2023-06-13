package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App    page.AppData
	Errors validation.List
	Form   *chooseAttorneysSummaryForm
	Lpa    *page.Lpa
}

func ChooseReplacementAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.ReplacementAttorneys) == 0 {
			return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys)
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: &chooseAttorneysSummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysSummaryForm(r, "yesToAddAnotherReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirectUrl string

				if data.Form.AddAttorney == "yes" {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseReplacementAttorneys)
				} else if len(lpa.ReplacementAttorneys) > 1 && (len(lpa.Attorneys) == 1 || lpa.AttorneyDecisions.How == actor.JointlyForSomeSeverallyForOthers || lpa.AttorneyDecisions.How == actor.Jointly) {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysMakeDecisions
				} else if lpa.AttorneyDecisions.How == actor.JointlyAndSeverally {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysStepIn
				} else {
					redirectUrl = page.Paths.TaskList
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}

		}

		return tmpl(w, data)
	}
}
