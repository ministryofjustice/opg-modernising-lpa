package page

import (
	"fmt"
	"net/http"
)

func DependencyHealthCheck(logger Logger, uidClient UidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := uidClient.Health(r.Context())

		if err != nil {
			logger.Print(fmt.Sprintf("Error while getting UID service status: %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(resp.StatusCode)
		}
	}
}
