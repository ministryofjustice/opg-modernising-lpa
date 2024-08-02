package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removePersonToNotifyData struct {
	App            appcontext.Data
	PersonToNotify donordata.PersonToNotify
	Errors         validation.List
	Form           *form.YesNoForm
}

func RemovePersonToNotify(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		person, found := donor.PeopleToNotify.Get(actoruid.FromRequest(r))

		if found == false {
			return page.Paths.ChoosePeopleToNotifySummary.Redirect(w, r, appData, donor)
		}

		data := &removePersonToNotifyData{
			App:            appData,
			PersonToNotify: person,
			Form:           form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveThisPerson")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					donor.PeopleToNotify.Delete(person)
					if len(donor.PeopleToNotify) == 0 {
						donor.Tasks.PeopleToNotify = task.StateNotStarted
					}

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return fmt.Errorf("error removing PersonToNotify from LPA: %w", err)
					}
				}

				return page.Paths.ChoosePeopleToNotifySummary.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
