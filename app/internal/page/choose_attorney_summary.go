package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseAttorneysSummaryData struct {
	App  AppData
	Form *chooseAttorneysSummaryForm
	Lpa  Lpa
}

type chooseAttorneysSummaryForm struct {
}

func ChooseAttorneySummary(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return nil
		}

		data := &chooseAttorneysSummaryData{
			App:  appData,
			Form: &chooseAttorneysSummaryForm{},
			Lpa:  lpa,
		}

		return tmpl(w, data)
	}
}
