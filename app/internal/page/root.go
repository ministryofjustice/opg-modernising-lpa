package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type rootData struct {
	App    AppData
	Errors validation.List
}

func Root(tmpl template.Template, logger Logger) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			http.Redirect(w, r, Paths.Start, http.StatusFound)
			return nil
		}

		w.WriteHeader(http.StatusNotFound)
		if terr := tmpl(w, &rootData{App: appData}); terr != nil {
			logger.Print(fmt.Sprintf("Error rendering page: %s", terr.Error()))
			http.Error(w, "Encountered an error", http.StatusInternalServerError)
		}

		return nil
	}
}
