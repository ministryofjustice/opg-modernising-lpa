package page

import (
	"net/http"
)

func DependencyHealthCheck(logger Logger, uidClient UidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := uidClient.Health(r.Context())

		if err != nil {
			logger.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if resp.StatusCode != http.StatusOK {
			logger.Print("UID service not available")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
