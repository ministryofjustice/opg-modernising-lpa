package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type doYouWantToNotifyPeopleData struct {
	App             AppData
	Errors          validation.List
	Form            *doYouWantToNotifyPeopleForm
	WantToNotify    string
	Lpa             *Lpa
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if len(lpa.PeopleToNotify) > 0 {
			return appData.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
		}

		data := &doYouWantToNotifyPeopleData{
			App:          appData,
			WantToNotify: lpa.DoYouWantToNotifyPeople,
			Lpa:          lpa,
		}

		switch lpa.HowAttorneysMakeDecisions {
		case Jointly:
			data.HowWorkTogether = "jointlyDescription"
		case JointlyAndSeverally:
			data.HowWorkTogether = "jointlyAndSeverallyDescription"
		case JointlyForSomeSeverallyForOthers:
			data.HowWorkTogether = "jointlyForSomeSeverallyForOthersDescription"
		}

		if r.Method == http.MethodPost {
			data.Form = readDoYouWantToNotifyPeople(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.DoYouWantToNotifyPeople = data.Form.WantToNotify
				lpa.Tasks.PeopleToNotify = TaskInProgress

				redirectPath := appData.Paths.ChoosePeopleToNotify

				if data.Form.WantToNotify == "no" {
					redirectPath = appData.Paths.CheckYourLpa
					lpa.Tasks.PeopleToNotify = TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
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
		WantToNotify: postFormString(r, "want-to-notify"),
	}
}

func (f *doYouWantToNotifyPeopleForm) Validate() validation.List {
	var errors validation.List

	errors.String("want-to-notify", "yesToNotifySomeoneAboutYourLpa", f.WantToNotify,
		validation.Select("yes", "no"))

	return errors
}
