package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysSummaryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func ChooseReplacementAttorneysSummary(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.ReplacementAttorneys.Len() == 0 {
			return donor.PathDoYouWantReplacementAttorneys.Redirect(w, r, appData, provided)
		}

		data := &chooseReplacementAttorneysSummaryData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherReplacementAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					return donor.PathChooseReplacementAttorneys.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else if provided.ReplacementAttorneys.Len() > 1 && (provided.Attorneys.Len() == 1 || provided.AttorneyDecisions.How.IsJointly()) {
					return donor.PathHowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, provided)
				} else if provided.AttorneyDecisions.How.IsJointlyAndSeverally() {
					return donor.PathHowShouldReplacementAttorneysStepIn.Redirect(w, r, appData, provided)
				} else {
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}

		}

		return tmpl(w, data)
	}
}
