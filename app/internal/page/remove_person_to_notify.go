package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removePersonToNotifyData struct {
	App            AppData
	PersonToNotify PersonToNotify
	Errors         validation.List
	Form           *removePersonToNotifyForm
}

func RemovePersonToNotify(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, found := lpa.GetPersonToNotify(id)

		if found == false {
			return appData.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
		}

		data := &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: attorney,
			Form:           &removePersonToNotifyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readRemovePersonToNotifyForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.RemovePersonToNotify == "yes" && data.Errors.Empty() {
				lpa.DeletePersonToNotify(attorney)

				var redirect string

				if len(lpa.PeopleToNotify) == 0 {
					lpa.Tasks.PeopleToNotify = TaskNotStarted
					redirect = appData.Paths.DoYouWantToNotifyPeople
				} else {
					redirect = appData.Paths.ChoosePeopleToNotifySummary
				}

				err = lpaStore.Put(r.Context(), lpa)

				if err != nil {
					logger.Print(fmt.Sprintf("error removing PersonToNotify from LPA: %s", err.Error()))
					return err
				}

				return appData.Redirect(w, r, lpa, redirect)
			}

			if data.Form.RemovePersonToNotify == "no" {
				return appData.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
			}

		}

		return tmpl(w, data)
	}
}

type removePersonToNotifyForm struct {
	RemovePersonToNotify string
}

func readRemovePersonToNotifyForm(r *http.Request) *removePersonToNotifyForm {
	return &removePersonToNotifyForm{
		RemovePersonToNotify: postFormString(r, "remove-person-to-notify"),
	}
}

func (f *removePersonToNotifyForm) Validate() validation.List {
	var errors validation.List

	if f.RemovePersonToNotify != "yes" && f.RemovePersonToNotify != "no" {
		errors.Add("remove-person-to-notify", "selectRemovePersonToNotify")
	}

	return errors
}
