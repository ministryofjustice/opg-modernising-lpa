package page

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type readYourLpaData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
	Json   string
}

func ReadYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &readYourLpaData{
			App: appData,
			Lpa: lpa,
		}

		b, err := json.Marshal(lpa)
		if err != nil {
			fmt.Println(err)
		}

		data.Json = string(b)

		return tmpl(w, data)
	}
}
