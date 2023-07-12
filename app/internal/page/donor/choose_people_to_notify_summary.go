package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Options form.YesNoOptions
	Lpa     *page.Lpa
}

func ChoosePeopleToNotifySummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.PeopleToNotify) == 0 {
			return appData.Redirect(w, r, lpa, page.Paths.DoYouWantToNotifyPeople.Format(lpa.ID))
		}

		data := &choosePeopleToNotifySummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &form.YesNoForm{},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherPersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := appData.Paths.ChoosePeopleToNotify.Format(lpa.ID) + "?addAnother=1"

				if data.Form.YesNo == form.No {
					redirectUrl = appData.Paths.TaskList.Format(lpa.ID)
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}
