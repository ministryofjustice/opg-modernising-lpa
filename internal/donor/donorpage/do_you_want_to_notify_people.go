package donorpage

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
	Form            *form.YesNoForm
	Donor           *actor.DonorProvidedDetails
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if len(donor.PeopleToNotify) > 0 {
			return page.Paths.ChoosePeopleToNotifySummary.Redirect(w, r, appData, donor)
		}

		data := &doYouWantToNotifyPeopleData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(donor.DoYouWantToNotifyPeople),
		}

		switch donor.AttorneyDecisions.How {
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
				donor.DoYouWantToNotifyPeople = data.Form.YesNo
				donor.Tasks.PeopleToNotify = actor.TaskInProgress

				redirectPath := page.Paths.ChoosePeopleToNotify

				if donor.DoYouWantToNotifyPeople == form.No {
					redirectPath = page.Paths.TaskList
					donor.Tasks.PeopleToNotify = actor.TaskCompleted
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirectPath.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
