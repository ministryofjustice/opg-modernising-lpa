package donor

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func Events() Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")

		// send a random number every 2 seconds
		for {
			rand.Seed(time.Now().UnixNano())
			fmt.Fprintf(w, "data: %d \n\n", rand.Intn(100))
			w.(http.Flusher).Flush()
			time.Sleep(2 * time.Second)
		}
	}
}
