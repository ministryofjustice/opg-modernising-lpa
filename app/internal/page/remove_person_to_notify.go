package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type removePersonToNotifyData struct {
	App            AppData
	PersonToNotify PersonToNotify
	Errors         map[string]string
	Form           removePersonToNotifyForm
}

type removePersonToNotifyForm struct {
	RemovePersonToNotify string
}

func RemovePersonToNotify(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, found := lpa.GetPersonToNotify(id)

		if found == false {
			return appData.Lang.Redirect(w, r, appData.Paths.ChoosePeopleToNotifySummary, http.StatusFound)
		}

		data := &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: attorney,
			Form:           removePersonToNotifyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = removePersonToNotifyForm{
				RemovePersonToNotify: postFormString(r, "remove-person-to-notify"),
			}

			data.Errors = data.Form.Validate()

			if data.Form.RemovePersonToNotify == "yes" && len(data.Errors) == 0 {
				lpa.DeletePersonToNotify(attorney)

				var redirect string

				if len(lpa.PeopleToNotify) == 0 {
					lpa.Tasks.PeopleToNotify = TaskNotStarted
					redirect = appData.Paths.DoYouWantToNotifyPeople
				} else {
					redirect = appData.Paths.ChoosePeopleToNotifySummary
				}

				err = lpaStore.Put(r.Context(), appData.SessionID, lpa)

				if err != nil {
					logger.Print(fmt.Sprintf("error removing PersonToNotify from LPA: %s", err.Error()))
					return err
				}

				return appData.Lang.Redirect(w, r, redirect, http.StatusFound)
			}

			if data.Form.RemovePersonToNotify == "no" {
				return appData.Lang.Redirect(w, r, appData.Paths.ChoosePeopleToNotifySummary, http.StatusFound)
			}

		}

		return tmpl(w, data)
	}
}

func (f *removePersonToNotifyForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.RemovePersonToNotify != "yes" && f.RemovePersonToNotify != "no" {
		errors["remove-person-to-notify"] = "selectRemovePersonToNotify"
	}

	return errors
}
