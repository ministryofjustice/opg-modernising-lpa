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

func ChoosePeopleToNotifySummary(logger page.Logger, tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
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
					redirectUrl = appData.Paths.CheckYourLpa
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
