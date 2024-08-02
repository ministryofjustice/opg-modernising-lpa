package page

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

func Recover(tmpl template.Template, logger Logger, bundle Bundle, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorContext(r.Context(), "recover error", slog.Any("req", r), slog.Any("err", err), slog.String("stack", string(debug.Stack())))
				w.WriteHeader(http.StatusInternalServerError)

				appData := appcontext.Data{CookieConsentSet: true}
				if strings.HasPrefix(r.URL.Path, "/cy/") {
					appData.Lang = localize.Cy
				} else {
					appData.Lang = localize.En
				}
				appData.Localizer = bundle.For(appData.Lang)

				if terr := tmpl(w, &errorData{App: appData}); terr != nil {
					logger.ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", terr))
					http.Error(w, "Encountered an error", http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}
}
