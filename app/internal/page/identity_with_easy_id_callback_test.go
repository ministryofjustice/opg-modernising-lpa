package page

import (
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

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithEasyIDCallbackData{
			App:      appData,
			FullName: "a-full-name",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithEasyIDCallback(template.Func, yotiClient)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, yotiClient, template)
}

func TestGetIdentityWithEasyIDCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithEasyIDCallback(nil, yotiClient)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, yotiClient)
}

func TestPostIdentityWithEasyIDCallback(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithEasyIDCallback(nil, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityOptionRedirectPath, resp.Header.Get("Location"))
}
