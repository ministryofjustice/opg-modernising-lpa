package page

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type recoverError struct {
	err   any
	stack []byte
}

func (e recoverError) Error() string { return "recover error" }
func (e recoverError) Title() string { return fmt.Sprint(e.err) }
func (e recoverError) Data() any     { return string(e.stack) }

func Recover(tmpl template.Template, logger Logger, bundle Bundle, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Request(r, recoverError{err: err, stack: debug.Stack()})
				w.WriteHeader(http.StatusInternalServerError)

				appData := AppData{CookieConsentSet: true}
				if strings.HasPrefix(r.URL.Path, "/cy/") {
					appData.Lang = localize.Cy
				} else {
					appData.Lang = localize.En
				}
				appData.Localizer = bundle.For(appData.Lang)

				if terr := tmpl(w, &errorData{App: appData}); terr != nil {
					logger.Print(fmt.Sprintf("Error rendering page: %s", terr.Error()))
					http.Error(w, "Encountered an error", http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}
}
