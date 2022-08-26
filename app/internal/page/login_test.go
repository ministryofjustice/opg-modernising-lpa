package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLoginClient struct {
	mock.Mock
}

func (m *mockLoginClient) AuthCodeURL(state, nonce, scope string) string {
	args := m.Called(state, nonce, scope)

	return args.String(0)
}

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "state-value", "nonce-value", "scope-value").
		Return("http://auth")

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Login(client)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, client)
}
