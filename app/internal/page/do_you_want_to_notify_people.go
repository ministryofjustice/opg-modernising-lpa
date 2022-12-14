package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type doYouWantToNotifyPeopleData struct {
	App          AppData
	Errors       map[string]string
	Form         *doYouWantToNotifyPeopleForm
	WantToNotify string
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

		data := &doYouWantToNotifyPeopleData{
			App:          appData,
			WantToNotify: lpa.DoYouWantToNotifyPeople,
		}

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

				appData.Lang.Redirect(w, r, redirectPath, http.StatusFound)

				return nil
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
