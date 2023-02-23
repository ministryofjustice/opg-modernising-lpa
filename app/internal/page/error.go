package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate mockery --testonly --inpackage --name ErrorHandler --structname mockErrorHandler
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type errorData struct {
	App    AppData
	Errors validation.List
}

func Error(tmpl template.Template, logger Logger) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Print(err)
		if err == ErrCsrfInvalid {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if terr := tmpl(w, &errorData{App: AppDataFromContext(r.Context())}); terr != nil {
			logger.Print(fmt.Sprintf("Error rendering page: %s", terr.Error()))
			http.Error(w, "Encountered an error", http.StatusInternalServerError)
		}
	}
}
