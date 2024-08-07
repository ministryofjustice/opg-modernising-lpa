package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func ChoosePeopleToNotifySummary(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) == 0 {
			return donor.PathDoYouWantToNotifyPeople.Redirect(w, r, appData, provided)
		}

		data := &choosePeopleToNotifySummaryData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherPersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.No {
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				} else {
					return donor.PathChoosePeopleToNotify.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}})
				}
			}
		}

		return tmpl(w, data)
	}
}
