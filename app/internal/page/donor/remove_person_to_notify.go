package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removePersonToNotifyData struct {
	App            page.AppData
	PersonToNotify actor.PersonToNotify
	Errors         validation.List
	Form           *removePersonToNotifyForm
}

func RemovePersonToNotify(logger Logger, tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		person, found := lpa.PeopleToNotify.Get(id)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary)
		}

		data := &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: person,
			Form:           &removePersonToNotifyForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readRemovePersonToNotifyForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.RemovePersonToNotify == "yes" {
					lpa.PeopleToNotify.Delete(person)
					if len(lpa.PeopleToNotify) == 0 {
						lpa.Tasks.PeopleToNotify = page.TaskNotStarted
					}

					if err := lpaStore.Put(r.Context(), lpa); err != nil {
						logger.Print(fmt.Sprintf("error removing PersonToNotify from LPA: %s", err.Error()))
						return err
					}
				}

				return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotifySummary)
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
		RemovePersonToNotify: page.PostFormString(r, "remove-person-to-notify"),
	}
}

func (f *removePersonToNotifyForm) Validate() validation.List {
	var errors validation.List

	errors.String("remove-person-to-notify", "removePersonToNotify", f.RemovePersonToNotify,
		validation.Select("yes", "no"))

	return errors
}
