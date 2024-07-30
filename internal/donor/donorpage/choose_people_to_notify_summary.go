package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *actor.DonorProvidedDetails
}

func ChoosePeopleToNotifySummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if len(donor.PeopleToNotify) == 0 {
			return page.Paths.DoYouWantToNotifyPeople.Redirect(w, r, appData, donor)
		}

		data := &choosePeopleToNotifySummaryData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherPersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.No {
					return page.Paths.TaskList.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.ChoosePeopleToNotify.RedirectQuery(w, r, appData, donor, url.Values{"addAnother": {"1"}})
				}
			}
		}

		return tmpl(w, data)
	}
}
