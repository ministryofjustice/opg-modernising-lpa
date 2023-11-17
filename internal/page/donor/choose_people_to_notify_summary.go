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

type choosePeopleToNotifySummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Options form.YesNoOptions
	Lpa     *actor.DonorProvidedDetails
}

func ChoosePeopleToNotifySummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		if len(lpa.PeopleToNotify) == 0 {
			return page.Paths.DoYouWantToNotifyPeople.Redirect(w, r, appData, lpa)
		}

		data := &choosePeopleToNotifySummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &form.YesNoForm{},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherPersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.No {
					return appData.Paths.TaskList.Redirect(w, r, appData, lpa)
				} else {
					return appData.Paths.ChoosePeopleToNotify.RedirectQuery(w, r, appData, lpa, url.Values{"addAnother": {"1"}})
				}
			}
		}

		return tmpl(w, data)
	}
}
