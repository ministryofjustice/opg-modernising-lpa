package page

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDependencyHealthCheck(t *testing.T) {
	testCases := map[int]struct {
		Body string
	}{
		200: {Body: "Its A-OK"},
		403: {Body: "Unauthorised"},
		503: {Body: "Something is wrong"},
	}

	for status, tc := range testCases {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		uidClient := newMockUidClient(t)
		uidClient.
			On("Health", mock.Anything).
			Return(&http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewBufferString(tc.Body))}, nil)

		DependencyHealthCheck(nil, uidClient)(w, r)

		resp := w.Result()
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		assert.Equal(t, status, resp.StatusCode)
		assert.Equal(t, tc.Body, string(body))
	}
}

func TestDependencyHealthCheckUidOnRequestError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	uidClient := newMockUidClient(t)
	uidClient.
		On("Health", mock.Anything).
		Return(&http.Response{}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError).
		Return(nil)

	DependencyHealthCheck(logger, uidClient)(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "Error while getting UID service status: err", string(body))
}
