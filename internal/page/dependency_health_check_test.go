package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDependencyHealthCheck(t *testing.T) {
	testCases := []int{
		200,
		403,
		503,
	}

	for _, status := range testCases {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		uidClient := newMockHealthChecker(t)
		uidClient.
			On("Health", mock.Anything).
			Return(&http.Response{StatusCode: status}, nil)

		DependencyHealthCheck(nil, uidClient)(w, r)

		resp := w.Result()

		assert.Equal(t, status, resp.StatusCode)
	}
}

func TestDependencyHealthCheckUidOnRequestError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	uidClient := newMockHealthChecker(t)
	uidClient.
		On("Health", mock.Anything).
		Return(&http.Response{}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("Error while getting UID service status: %s", expectedError)).
		Return(nil)

	DependencyHealthCheck(logger, uidClient)(w, r)

	resp := w.Result()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
