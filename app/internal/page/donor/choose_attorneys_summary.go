package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type chooseAttorneysSummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *chooseAttorneysSummaryForm
	Lpa     *page.Lpa
	Options actor.YesNoOptions
}

func ChooseAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.Attorneys) == 0 {
			return appData.Redirect(w, r, lpa, fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseAttorneys.Format(lpa.ID)))
		}

		data := &chooseAttorneysSummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &chooseAttorneysSummaryForm{},
			Options: actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysSummaryForm(r, "yesToAddAnotherAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := appData.Paths.TaskList.Format(lpa.ID)

				if len(lpa.Attorneys) > 1 {
					redirectUrl = appData.Paths.HowShouldAttorneysMakeDecisions.Format(lpa.ID)
				}

				if data.Form.AddAttorney == actor.Yes {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseAttorneys.Format(lpa.ID))
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysSummaryForm struct {
	AddAttorney actor.YesNo
	Error       error
	errorLabel  string
}

func readChooseAttorneysSummaryForm(r *http.Request, errorLabel string) *chooseAttorneysSummaryForm {
	add, err := actor.ParseYesNo(page.PostFormString(r, "add-attorney"))

	return &chooseAttorneysSummaryForm{
		AddAttorney: add,
		Error:       err,
		errorLabel:  errorLabel,
	}
}

func (f *chooseAttorneysSummaryForm) Validate() validation.List {
	var errors validation.List

	errors.Error("add-attorney", f.errorLabel, f.Error,
		validation.Selected())

	return errors
}
