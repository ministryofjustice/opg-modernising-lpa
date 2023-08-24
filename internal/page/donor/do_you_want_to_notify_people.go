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
	Lpa             *page.Lpa
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.PeopleToNotify) > 0 {
			return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary.Format(lpa.ID))
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

				redirectPath := appData.Paths.ChoosePeopleToNotify.Format(lpa.ID)

				if lpa.DoYouWantToNotifyPeople == form.No {
					redirectPath = appData.Paths.TaskList.Format(lpa.ID)
					lpa.Tasks.PeopleToNotify = actor.TaskCompleted
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirectPath)
			}
		}

		return tmpl(w, data)
	}
}
