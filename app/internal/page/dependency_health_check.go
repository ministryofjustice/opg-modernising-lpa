package page

import (
	"io"
	"net/http"
)

func DependencyHealthCheck(logger Logger, uidClient UidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := uidClient.Health(r.Context())

		if err != nil {
			logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(resp.Body)

		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}
