package donor

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Lpa     *actor.DonorProvidedDetails
	Options form.YesNoOptions
}

func ChooseReplacementAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		if lpa.ReplacementAttorneys.Len() == 0 {
			return page.Paths.DoYouWantReplacementAttorneys.Redirect(w, r, appData, lpa)
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
				if data.Form.YesNo == form.Yes {
					return appData.Paths.ChooseReplacementAttorneys.RedirectQuery(w, r, appData, lpa, url.Values{"addAnother": {"1"}})
				} else if lpa.ReplacementAttorneys.Len() > 1 && (lpa.Attorneys.Len() == 1 || lpa.AttorneyDecisions.How.IsJointlyForSomeSeverallyForOthers() || lpa.AttorneyDecisions.How.IsJointly()) {
					return appData.Paths.HowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, lpa)
				} else if lpa.AttorneyDecisions.How.IsJointlyAndSeverally() {
					return appData.Paths.HowShouldReplacementAttorneysStepIn.Redirect(w, r, appData, lpa)
				} else {
					return page.Paths.TaskList.Redirect(w, r, appData, lpa)
				}
			}

		}

		return tmpl(w, data)
	}
}
