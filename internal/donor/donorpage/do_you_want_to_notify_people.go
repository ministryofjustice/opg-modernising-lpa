package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type doYouWantToNotifyPeopleData struct {
	App             appcontext.Data
	Errors          validation.List
	Form            *form.YesNoForm
	Donor           *donordata.Provided
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if len(donor.PeopleToNotify) > 0 {
			return page.Paths.ChoosePeopleToNotifySummary.Redirect(w, r, appData, donor)
		}

		data := &doYouWantToNotifyPeopleData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(donor.DoYouWantToNotifyPeople),
		}

		switch donor.AttorneyDecisions.How {
		case lpadata.Jointly:
			data.HowWorkTogether = "jointlyDescription"
		case lpadata.JointlyAndSeverally:
			data.HowWorkTogether = "jointlyAndSeverallyDescription"
		case lpadata.JointlyForSomeSeverallyForOthers:
			data.HowWorkTogether = "jointlyForSomeSeverallyForOthersDescription"
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToNotifySomeoneAboutYourLpa")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.DoYouWantToNotifyPeople = data.Form.YesNo
				donor.Tasks.PeopleToNotify = task.StateInProgress

				redirectPath := page.Paths.ChoosePeopleToNotify

				if donor.DoYouWantToNotifyPeople == form.No {
					redirectPath = page.Paths.TaskList
					donor.Tasks.PeopleToNotify = task.StateCompleted
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
