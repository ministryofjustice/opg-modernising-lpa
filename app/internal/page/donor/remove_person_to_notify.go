package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type removePersonToNotifyData struct {
	App            page.AppData
	PersonToNotify actor.PersonToNotify
	Errors         validation.List
	Form           *removePersonToNotifyForm
	Options        actor.YesNoOptions
}

func RemovePersonToNotify(logger Logger, tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		id := r.FormValue("id")
		person, found := lpa.PeopleToNotify.Get(id)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary.Format(lpa.ID))
		}

		data := &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: person,
			Form:           &removePersonToNotifyForm{},
			Options:        actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readRemovePersonToNotifyForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.RemovePersonToNotify == actor.Yes {
					lpa.PeopleToNotify.Delete(person)
					if len(lpa.PeopleToNotify) == 0 {
						lpa.Tasks.PeopleToNotify = actor.TaskNotStarted
					}

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						logger.Print(fmt.Sprintf("error removing PersonToNotify from LPA: %s", err.Error()))
						return err
					}
				}

				return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type removePersonToNotifyForm struct {
	RemovePersonToNotify actor.YesNo
	Error                error
}

func readRemovePersonToNotifyForm(r *http.Request) *removePersonToNotifyForm {
	remove, err := actor.ParseYesNo(page.PostFormString(r, "remove-person-to-notify"))

	return &removePersonToNotifyForm{
		RemovePersonToNotify: remove,
		Error:                err,
	}
}

func (f *removePersonToNotifyForm) Validate() validation.List {
	var errors validation.List

	errors.Error("remove-person-to-notify", "removePersonToNotify", f.Error,
		validation.Selected())

	return errors
}
