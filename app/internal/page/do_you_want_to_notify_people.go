package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type doYouWantToNotifyPeopleData struct {
	App             AppData
	Errors          map[string]string
	Form            *doYouWantToNotifyPeopleForm
	WantToNotify    string
	Lpa             *Lpa
	HowWorkTogether string
}

type doYouWantToNotifyPeopleForm struct {
	WantToNotify string
}

func DoYouWantToNotifyPeople(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if len(lpa.PeopleToNotify) > 0 {
			return appData.Lang.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
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

		//Jointly - that you want your attorneys to work together to make decisions
		//Jointly and Severally - that you want your attorneys to work together or separately to make decisions
		//Jointly for some and severally for others - that you want your attorneys to work together on some decisions but can make other decisions separately

		if r.Method == http.MethodPost {
			data.Form = readDoYouWantToNotifyPeople(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.DoYouWantToNotifyPeople = data.Form.WantToNotify
				lpa.Tasks.PeopleToNotify = TaskInProgress

				redirectPath := appData.Paths.ChoosePeopleToNotify

				if data.Form.WantToNotify == "no" {
					redirectPath = appData.Paths.CheckYourLpa
					lpa.Tasks.PeopleToNotify = TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, redirectPath)
			}
		}

		return tmpl(w, data)
	}
}

func readDoYouWantToNotifyPeople(r *http.Request) *doYouWantToNotifyPeopleForm {
	r.ParseForm()

	return &doYouWantToNotifyPeopleForm{
		WantToNotify: postFormString(r, "want-to-notify"),
	}
}

func (f *doYouWantToNotifyPeopleForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.WantToNotify != "yes" && f.WantToNotify != "no" {
		errors["want-to-notify"] = "selectDoYouWantToNotifyPeople"
	}

	return errors
}
