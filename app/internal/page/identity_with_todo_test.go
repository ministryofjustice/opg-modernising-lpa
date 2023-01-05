package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithTodo(t *testing.T) {
	w := httptest.NewRecorder()

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithTodoData{
			App:            appData,
			IdentityOption: Passport,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithTodo(template.Func, Passport)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostIdentityWithTodo(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithTodo(nil, Passport)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.WitnessingAsCertificateProvider, resp.Header.Get("Location"))
}
