package donorpage

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func UploadEvidenceSSE(documentStore DocumentStore, ttl time.Duration, flushFrequency time.Duration, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")

		documents, err := documentStore.GetAll(r.Context())
		if err != nil {
			printMessage("data: {\"closeConnection\": \"1\"}\n\n", w)
			return nil
		}

		alreadyScannedCount := len(documents.Scanned())
		batchToBeScannedCount := len(documents.NotScanned())

		for start := now(); time.Since(start) < ttl; {
			documents, err := documentStore.GetAll(r.Context())
			if err != nil {
				printMessage("data: {\"closeConnection\": \"1\"}\n\n", w)
				return nil
			}

			scannedCount := len(documents.Scanned()) - alreadyScannedCount

			printMessage(fmt.Sprintf("data: {\"finishedScanning\": %v, \"scannedCount\": %d}\n\n", scannedCount == batchToBeScannedCount, scannedCount), w)

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
