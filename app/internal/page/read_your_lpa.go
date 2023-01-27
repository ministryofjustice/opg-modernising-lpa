package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readYourLpaData struct {
	App    AppData
	Errors validation.List
	Lpa    *Lpa
}

func ReadYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &readYourLpaData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
