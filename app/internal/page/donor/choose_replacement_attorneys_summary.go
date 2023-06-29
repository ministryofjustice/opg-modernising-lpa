package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *chooseAttorneysSummaryForm
	Lpa     *page.Lpa
	Options actor.YesNoOptions
}

func ChooseReplacementAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.ReplacementAttorneys) == 0 {
			return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys.Format(lpa.ID))
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &chooseAttorneysSummaryForm{},
			Options: actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysSummaryForm(r, "yesToAddAnotherReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirectUrl string

				if data.Form.AddAttorney == actor.Yes {
					redirectUrl = appData.Paths.ChooseReplacementAttorneys.Format(lpa.ID) + "?addAnother=1"
				} else if len(lpa.ReplacementAttorneys) > 1 && (len(lpa.Attorneys) == 1 || lpa.AttorneyDecisions.How == actor.JointlyForSomeSeverallyForOthers || lpa.AttorneyDecisions.How == actor.Jointly) {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysMakeDecisions.Format(lpa.ID)
				} else if lpa.AttorneyDecisions.How == actor.JointlyAndSeverally {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysStepIn.Format(lpa.ID)
				} else {
					redirectUrl = page.Paths.TaskList.Format(lpa.ID)
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}

		}

		return tmpl(w, data)
	}
}
