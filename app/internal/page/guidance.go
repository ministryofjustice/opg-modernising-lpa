package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type guidanceData struct {
	App      AppData
	Errors   map[string]string
	Continue string
	Lpa      Lpa
}

func Guidance(tmpl template.Template, continuePath string, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &guidanceData{
			App:      appData,
			Continue: continuePath,
		}

		if dataStore != nil {
			if err := dataStore.Get(r.Context(), appData.SessionID, &data.Lpa); err != nil {
				return err
			}
		}

		return tmpl(w, data)
	}
}
