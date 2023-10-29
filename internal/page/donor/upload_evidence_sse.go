package donor

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func UploadEvidenceSSE(store DonorStore, ttl time.Duration, flushFrequency time.Duration) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")

		fileTotal := len(lpa.Evidence.Documents)

		for start := time.Now(); time.Since(start) < ttl; {
			lpa, err := store.Get(r.Context())
			if err != nil {
				printMessage("data: {\"closeConnection\": \"1\"}\n\n", w)
				return err
			}

			printMessage(fmt.Sprintf("data: {\"fileTotal\": %d, \"scannedTotal\": %d}\n\n", fileTotal, lpa.Evidence.ScannedCount()), w)

			time.Sleep(flushFrequency)
		}

		printMessage("data: {\"closeConnection\": \"1\"}\n\n", w)

		return nil
	}
}

func printMessage(message string, w io.Writer) {
	fmt.Fprint(w, "event: message\n")
	fmt.Fprint(w, message)
	w.(http.Flusher).Flush()
}
