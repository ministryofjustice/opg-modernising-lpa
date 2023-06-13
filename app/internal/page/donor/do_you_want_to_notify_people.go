package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type doYouWantToNotifyPeopleData struct {
	App             page.AppData
	Errors          validation.List
	Form            *doYouWantToNotifyPeopleForm
	WantToNotify    string
	Lpa             *page.Lpa
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.PeopleToNotify) > 0 {
			return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary)
		}

		data := &doYouWantToNotifyPeopleData{
			App:          appData,
			WantToNotify: lpa.DoYouWantToNotifyPeople,
			Lpa:          lpa,
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
			data.Form = readDoYouWantToNotifyPeople(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.DoYouWantToNotifyPeople = data.Form.WantToNotify
				lpa.Tasks.PeopleToNotify = actor.TaskInProgress

				redirectPath := appData.Paths.ChoosePeopleToNotify

				if data.Form.WantToNotify == "no" {
					redirectPath = appData.Paths.TaskList
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

type doYouWantToNotifyPeopleForm struct {
	WantToNotify string
}

func readDoYouWantToNotifyPeople(r *http.Request) *doYouWantToNotifyPeopleForm {
	return &doYouWantToNotifyPeopleForm{
		WantToNotify: page.PostFormString(r, "want-to-notify"),
	}
}

func (f *doYouWantToNotifyPeopleForm) Validate() validation.List {
	var errors validation.List

	errors.String("want-to-notify", "yesToNotifySomeoneAboutYourLpa", f.WantToNotify,
		validation.Select("yes", "no"))

	return errors
}
