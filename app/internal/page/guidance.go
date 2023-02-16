package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    AppData
	Errors validation.List
	Lpa    *Lpa
}

func Guidance(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &guidanceData{
			App: appData,
		}

		if lpaStore != nil {
			lpa, err := lpaStore.Get(r.Context())
			if err != nil {
				return err
			}
			data.Lpa = lpa
		}

		return tmpl(w, data)
	}
}
