package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type removePersonToNotifyData struct {
	App            page.AppData
	PersonToNotify actor.PersonToNotify
	Errors         validation.List
	Form           *form.YesNoForm
	Options        form.YesNoOptions
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
			Form:           &form.YesNoForm{},
			Options:        form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "removePersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
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
