package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseAttorneysSummaryData struct {
	App                 AppData
	AttorneyAddressPath string
	AttorneyDetailsPath string
	Errors              map[string]string
	Lpa                 Lpa
	Form                chooseAttorneysSummaryForm
}

type chooseAttorneysSummaryForm struct {
	AddAttorney string
}

func ChooseAttorneySummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseAttorneysSummaryData{
			App:                 appData,
			Lpa:                 lpa,
			AttorneyDetailsPath: chooseAttorneysPath,
			AttorneyAddressPath: chooseAttorneysAddressPath,
			Form:                chooseAttorneysSummaryForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = chooseAttorneysSummaryForm{
				AddAttorney: postFormString(r, "add-attorney"),
			}

			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				redirectUrl := wantReplacementAttorneysPath

				if data.Form.AddAttorney == "yes" {
					redirectUrl = data.AttorneyDetailsPath
				}

				appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
				return nil
			}

		}

		return tmpl(w, data)
	}
}

func (f *chooseAttorneysSummaryForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.AddAttorney != "yes" && f.AddAttorney != "no" {
		errors = map[string]string{
			"add-attorney": "selectAddMoreAttorneys",
		}
	}

	return errors
}
