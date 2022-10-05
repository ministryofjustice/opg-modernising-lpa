package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{FullName: "a-full-name"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:      appData,
			FullName: "a-full-name",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(template.Func, yotiClient)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, yotiClient, template)
}

func TestGetIdentityWithYotiCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, yotiClient)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, yotiClient)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithYotiCallback(nil, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityOptionRedirectPath, resp.Header.Get("Location"))
}
