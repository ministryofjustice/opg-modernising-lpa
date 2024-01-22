package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func TODO() Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("<!doctype HTML><p>TODO</p>"))
		return err
	}
}
