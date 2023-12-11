package page

import (
	"context"
	"encoding/json"
	"net/http"
)

//go:generate mockery --testonly --inpackage --name HealthChecker --structname mockHealthChecker
type HealthChecker interface {
	CheckHealth(context.Context) error
}

func DependencyHealthCheck(logger Logger, services map[string]HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := map[string]string{}
		status := http.StatusOK

		for name, service := range services {
			if err := service.CheckHealth(r.Context()); err != nil {
				status = http.StatusBadRequest
				results[name] = err.Error()
			}
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(results)
	}
}
