package donor

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func UploadEvidenceSSE(store DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")

		fileTotal := len(lpa.Evidence.Documents)

		for {
			lpa, err := store.Get(r.Context())
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "data: {\"fileTotal\": %d, \"scannedTotal\": %d} \n\n", fileTotal, lpa.Evidence.ScannedCount())
			w.(http.Flusher).Flush()

			time.Sleep(2 * time.Second)
		}
	}
}
