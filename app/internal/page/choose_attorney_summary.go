package page

import (
	"fmt"
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
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		data := &chooseAttorneysSummaryData{
			App:  appData,
			Form: &chooseAttorneysSummaryForm{},
			Lpa:  lpa,
		}

		return tmpl(w, data)
	}
}
