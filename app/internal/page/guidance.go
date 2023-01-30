package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App      AppData
	Errors   validation.List
	Continue string
	Lpa      *Lpa
}

func Guidance(tmpl template.Template, continuePath string, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if continuePath == appData.Paths.Auth {
			continuePath = fmt.Sprintf("%s?locale=%s", continuePath, appData.Lang.String())
		}

		data := &guidanceData{
			App:      appData,
			Continue: continuePath,
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
