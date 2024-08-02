package page

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type errorData struct {
	App    appcontext.Data
	Errors validation.List
	Err    error
}

func Error(tmpl template.Template, logger Logger, showError bool) ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.ErrorContext(r.Context(), "request error", slog.Any("req", r), slog.Any("err", err))
		if err == ErrCsrfInvalid {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		data := &errorData{App: appcontext.DataFromContext(r.Context())}
		if showError {
			data.Err = err
		}

		if terr := tmpl(w, data); terr != nil {
			logger.ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", terr))
			http.Error(w, "Encountered an error", http.StatusInternalServerError)
		}
	}
}
