package page

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type rootData struct {
	App    appcontext.Data
	Errors validation.List
}

func Root(tmpl template.Template, logger Logger) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path == "/" {
			http.Redirect(w, r, Paths.Start.Format(), http.StatusFound)
			return nil
		}

		w.WriteHeader(http.StatusNotFound)
		if terr := tmpl(w, &rootData{App: appData}); terr != nil {
			logger.ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", terr))
			http.Error(w, "Encountered an error", http.StatusInternalServerError)
		}

		return nil
	}
}
