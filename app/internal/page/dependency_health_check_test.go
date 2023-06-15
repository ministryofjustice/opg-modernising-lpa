package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDependencyHealthCheck(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", mock.Anything, mock.Anything).
		Return(uid.CreateCaseResponse{}, nil)

	DependencyHealthCheck(nil, uidClient)(w, r)

	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDependencyHealthCheckUidUnhealthy(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", mock.Anything, mock.Anything).
		Return(uid.CreateCaseResponse{}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError).
		Return(nil)

	DependencyHealthCheck(logger, uidClient)(w, r)

	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
