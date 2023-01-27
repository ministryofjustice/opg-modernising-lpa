package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysSummaryData struct {
	App    AppData
	Errors validation.List
	Form   chooseAttorneysSummaryForm
	Lpa    *Lpa
}

func ChooseAttorneysSummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseAttorneysSummaryData{
			App:  appData,
			Lpa:  lpa,
			Form: chooseAttorneysSummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = *readChooseAttorneysSummaryForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.Empty() {
				redirectUrl := appData.Paths.DoYouWantReplacementAttorneys

				if len(lpa.Attorneys) > 1 {
					redirectUrl = appData.Paths.HowShouldAttorneysMakeDecisions
				}

				if data.Form.AddAttorney == "yes" {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseAttorneys)
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}

		}

		return tmpl(w, data)
	}
}

type chooseAttorneysSummaryForm struct {
	AddAttorney string
}

func readChooseAttorneysSummaryForm(r *http.Request) *chooseAttorneysSummaryForm {
	return &chooseAttorneysSummaryForm{
		AddAttorney: postFormString(r, "add-attorney"),
	}
}

func (f *chooseAttorneysSummaryForm) Validate() validation.List {
	var errors validation.List

	if f.AddAttorney != "yes" && f.AddAttorney != "no" {
		errors.Add("add-attorney", "selectAddMoreAttorneys")
	}

	return errors
}
