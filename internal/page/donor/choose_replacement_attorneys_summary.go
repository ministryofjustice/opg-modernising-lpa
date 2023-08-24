package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Lpa     *page.Lpa
	Options form.YesNoOptions
}

func ChooseReplacementAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if lpa.ReplacementAttorneys.Len() == 0 {
			return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys.Format(lpa.ID))
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &form.YesNoForm{},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirectUrl string

				if data.Form.YesNo == form.Yes {
					redirectUrl = appData.Paths.ChooseReplacementAttorneys.Format(lpa.ID) + "?addAnother=1"
				} else if lpa.ReplacementAttorneys.Len() > 1 && (lpa.Attorneys.Len() == 1 || lpa.AttorneyDecisions.How.IsJointlyForSomeSeverallyForOthers() || lpa.AttorneyDecisions.How.IsJointly()) {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysMakeDecisions.Format(lpa.ID)
				} else if lpa.AttorneyDecisions.How.IsJointlyAndSeverally() {
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
