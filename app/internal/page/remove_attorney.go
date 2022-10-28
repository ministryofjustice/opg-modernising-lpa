package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type removeAttorneyData struct {
	App      AppData
	Attorney Attorney
	Errors   map[string]string
	Form     removeAttorneyForm
}

type removeAttorneyForm struct {
	RemoveAttorney string
}

func RemoveAttorney(logger Logger, tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			logger.Print(fmt.Sprintf("error getting lpa from store: %s", err.Error()))
			return err
		}

		id := r.FormValue("id")
		attorney, _ := lpa.GetAttorney(id)

		data := &removeAttorneyData{
			App:      appData,
			Attorney: attorney,
			Form:     removeAttorneyForm{},
		}

		return tmpl(w, data)
	}
}
