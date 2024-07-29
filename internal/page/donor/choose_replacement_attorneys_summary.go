package donor

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *actor.DonorProvidedDetails
}

func ChooseReplacementAttorneysSummary(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if donor.ReplacementAttorneys.Len() == 0 {
			return page.Paths.DoYouWantReplacementAttorneys.Redirect(w, r, appData, donor)
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					return page.Paths.ChooseReplacementAttorneys.RedirectQuery(w, r, appData, donor, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else if donor.ReplacementAttorneys.Len() > 1 && (donor.Attorneys.Len() == 1 || donor.AttorneyDecisions.How.IsJointly()) {
					return page.Paths.HowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, donor)
				} else if donor.AttorneyDecisions.How.IsJointlyAndSeverally() {
					return page.Paths.HowShouldReplacementAttorneysStepIn.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.TaskList.Redirect(w, r, appData, donor)
				}
			}

		}

		return tmpl(w, data)
	}
}
