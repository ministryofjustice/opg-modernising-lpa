package page

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithEasyIDCallback(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{FullName: "a-full-name"}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithEasyIDCallback(yotiClient)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "<!doctype html><p>Hi a-full-name</p>", string(data))

	mock.AssertExpectationsForObjects(t, yotiClient)
}

func TestGetIdentityWithEasyIDCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithEasyIDCallback(yotiClient)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, yotiClient)
}
