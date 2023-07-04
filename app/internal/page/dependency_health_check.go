package page

import (
	"fmt"
	"io"
	"net/http"
)

func DependencyHealthCheck(logger Logger, uidClient UidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := uidClient.Health(r.Context())

		if err != nil {
			logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Error while getting UID service status: %s", err.Error())))
		} else {
			resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			w.WriteHeader(resp.StatusCode)
			w.Write(body)
		}
	}
}
