package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howShouldAttorneysMakeDecisionsData struct {
	App              AppData
	DecisionsDetails string
	Errors           map[string]string
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			fmt.Print(lpa.ID)
			return err
		}

		data := &howShouldAttorneysMakeDecisionsData{
			App:              appData,
			DecisionsDetails: lpa.DecisionsDetails,
		}

		return tmpl(w, data)
	}
}
