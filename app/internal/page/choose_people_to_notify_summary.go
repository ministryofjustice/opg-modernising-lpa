package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type choosePeopleToNotifySummaryData struct {
	App    AppData
	Errors map[string]string
	Form   choosePeopleToNotifySummaryForm
	Lpa    *Lpa
}

type choosePeopleToNotifySummaryForm struct {
	AddPersonToNotify string
}

func ChoosePeopleToNotifySummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &choosePeopleToNotifySummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: choosePeopleToNotifySummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = choosePeopleToNotifySummaryForm{
				AddPersonToNotify: postFormString(r, "add-person-to-notify"),
			}

			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				redirectUrl := fmt.Sprintf("%s?addAnother=1", appData.Paths.ChoosePeopleToNotify)

				if data.Form.AddPersonToNotify == "no" {
					redirectUrl = appData.Paths.CheckYourLpa
					lpa.Tasks.PeopleToNotify = TaskCompleted

					if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
						return err
					}
				}

				return appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
			}

		}

		return tmpl(w, data)
	}
}

func (f *choosePeopleToNotifySummaryForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.AddPersonToNotify != "yes" && f.AddPersonToNotify != "no" {
		errors = map[string]string{
			"add-person-to-notify": "selectAddMorePeopleToNotify",
		}
	}

	return errors
}
