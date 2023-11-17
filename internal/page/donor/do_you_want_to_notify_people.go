package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type doYouWantToNotifyPeopleData struct {
	App             page.AppData
	Errors          validation.List
	Options         form.YesNoOptions
	Form            *form.YesNoForm
	Lpa             *actor.DonorProvidedDetails
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.DonorProvidedDetails) error {
		if len(lpa.PeopleToNotify) > 0 {
			return page.Paths.ChoosePeopleToNotifySummary.Redirect(w, r, appData, lpa)
		}

		data := &doYouWantToNotifyPeopleData{
			App: appData,
			Lpa: lpa,
			Form: &form.YesNoForm{
				YesNo: lpa.DoYouWantToNotifyPeople,
			},
			Options: form.YesNoValues,
		}

		switch lpa.AttorneyDecisions.How {
		case actor.Jointly:
			data.HowWorkTogether = "jointlyDescription"
		case actor.JointlyAndSeverally:
			data.HowWorkTogether = "jointlyAndSeverallyDescription"
		case actor.JointlyForSomeSeverallyForOthers:
			data.HowWorkTogether = "jointlyForSomeSeverallyForOthersDescription"
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToNotifySomeoneAboutYourLpa")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.DoYouWantToNotifyPeople = data.Form.YesNo
				lpa.Tasks.PeopleToNotify = actor.TaskInProgress

				redirectPath := appData.Paths.ChoosePeopleToNotify

				if lpa.DoYouWantToNotifyPeople == form.No {
					redirectPath = appData.Paths.TaskList
					lpa.Tasks.PeopleToNotify = actor.TaskCompleted
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return redirectPath.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}
