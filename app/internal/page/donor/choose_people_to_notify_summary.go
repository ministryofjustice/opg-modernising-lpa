package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *choosePeopleToNotifySummaryForm
	Options actor.YesNoOptions
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
			Form:    &choosePeopleToNotifySummaryForm{},
			Options: actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readChoosePeopleToNotifySummaryForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := appData.Paths.ChoosePeopleToNotify.Format(lpa.ID) + "?addAnother=1"

				if data.Form.AddPersonToNotify == actor.No {
					redirectUrl = appData.Paths.TaskList.Format(lpa.ID)
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}

type choosePeopleToNotifySummaryForm struct {
	AddPersonToNotify actor.YesNo
	Error             error
}

func readChoosePeopleToNotifySummaryForm(r *http.Request) *choosePeopleToNotifySummaryForm {
	add, err := actor.ParseYesNo(page.PostFormString(r, "add-person-to-notify"))

	return &choosePeopleToNotifySummaryForm{
		AddPersonToNotify: add,
		Error:             err,
	}
}

func (f *choosePeopleToNotifySummaryForm) Validate() validation.List {
	var errors validation.List

	errors.Error("add-person-to-notify", "yesToAddAnotherPersonToNotify", f.Error,
		validation.Selected())

	return errors
}
