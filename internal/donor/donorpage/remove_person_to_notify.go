package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removePersonToNotifyData struct {
	App            appcontext.Data
	PersonToNotify donordata.PersonToNotify
	Errors         validation.List
	Form           *form.YesNoForm
}

func RemovePersonToNotify(tmpl template.Template, service PeopleToNotifyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		person, found := provided.PeopleToNotify.Get(actoruid.FromRequest(r))

		if found == false {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
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
					if err := service.Delete(r.Context(), provided, person); err != nil {
						return err
					}
				}

				return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
