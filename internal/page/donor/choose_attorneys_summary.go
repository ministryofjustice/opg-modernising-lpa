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

type chooseAttorneysSummaryData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *actor.DonorProvidedDetails
}

func ChooseAttorneysSummary(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if donor.Attorneys.Len() == 0 {
			return page.Paths.ChooseAttorneys.RedirectQuery(w, r, appData, donor, url.Values{"addAnother": {"1"}})
		}

		data := &chooseAttorneysSummaryData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := page.Paths.TaskList
				if donor.Attorneys.Len() > 1 {
					redirectUrl = page.Paths.HowShouldAttorneysMakeDecisions
				}

				if data.Form.YesNo.IsYes() {
					return page.Paths.ChooseAttorneys.RedirectQuery(w, r, appData, donor, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else {
					return redirectUrl.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}
