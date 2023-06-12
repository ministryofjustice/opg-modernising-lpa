package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App    page.AppData
	Errors validation.List
	Form   *choosePeopleToNotifySummaryForm
	Lpa    *page.Lpa
}

func ChoosePeopleToNotifySummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.PeopleToNotify) == 0 {
			return appData.Redirect(w, r, lpa, page.Paths.DoYouWantToNotifyPeople)
		}

		data := &choosePeopleToNotifySummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: &choosePeopleToNotifySummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readChoosePeopleToNotifySummaryForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := fmt.Sprintf("%s?addAnother=1", appData.Paths.ChoosePeopleToNotify)

				if data.Form.AddPersonToNotify == "no" {
					redirectUrl = appData.Paths.TaskList
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}

type choosePeopleToNotifySummaryForm struct {
	AddPersonToNotify string
}

func readChoosePeopleToNotifySummaryForm(r *http.Request) *choosePeopleToNotifySummaryForm {
	return &choosePeopleToNotifySummaryForm{
		AddPersonToNotify: page.PostFormString(r, "add-person-to-notify"),
	}
}

func (f *choosePeopleToNotifySummaryForm) Validate() validation.List {
	var errors validation.List

	errors.String("add-person-to-notify", "yesToAddAnotherPersonToNotify", f.AddPersonToNotify,
		validation.Select("yes", "no"))

	return errors
}
