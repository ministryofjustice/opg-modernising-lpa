package page

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyHealthCheck(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	service := newMockHealthChecker(t)
	service.EXPECT().CheckHealth(r.Context()).Return(nil)

	services := map[string]HealthChecker{
		"service": service,
	}

	DependencyHealthCheck(nil, services)(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "{}\n", string(body))
}

func TestDependencyHealthCheckWhenError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	service := newMockHealthChecker(t)
	service.EXPECT().CheckHealth(r.Context()).Return(nil)

	badService := newMockHealthChecker(t)
	badService.EXPECT().CheckHealth(r.Context()).Return(expectedError)

	services := map[string]HealthChecker{
		"service":    service,
		"badService": badService,
	}

	DependencyHealthCheck(nil, services)(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, fmt.Sprintf("{\"badService\":\"%s\"}\n", expectedError.Error()), string(body))
}
