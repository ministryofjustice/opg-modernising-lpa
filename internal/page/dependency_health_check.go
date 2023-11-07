package page

import (
	"context"
	"fmt"
	"net/http"
)

//go:generate mockery --testonly --inpackage --name HealthChecker --structname mockHealthChecker
type HealthChecker interface {
	Health(context.Context) (*http.Response, error)
}

func DependencyHealthCheck(logger Logger, uidService HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := uidService.Health(r.Context())

		if err != nil {
			logger.Print(fmt.Sprintf("Error while getting UID service status: %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(resp.StatusCode)
		}
	}
}
