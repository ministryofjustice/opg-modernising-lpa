package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type errorData struct {
	App    AppData
	Errors validation.List
}

func Error(tmpl template.Template, logger Logger) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Request(r, err)
		if err == ErrCsrfInvalid {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if terr := tmpl(w, &errorData{App: AppDataFromContext(r.Context())}); terr != nil {
			logger.Request(r, fmt.Errorf("Error rendering page: %w", terr))
			http.Error(w, "Encountered an error", http.StatusInternalServerError)
		}
	}
}
