package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSignInClient struct {
	signin.ClientInterface
	mock.Mock
}

func (m *mockSignInClient) GetToken(redirectUri, clientID, JTI, code string) (string, error) {
	args := m.Called(redirectUri, clientID, JTI, code)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockSignInClient) GetUserInfo(jwt string) (signin.UserInfoResponse, error) {
	args := m.Called(jwt)
	return args.Get(0).(signin.UserInfoResponse), args.Error(1)
}

func TestSignInCallback(t *testing.T) {
	w := httptest.NewRecorder()

	c := &mockSignInClient{}
	c.
		On("GetToken", "/app/public-url:/home", "client-id", "jti", "auth-code").
		Return("a JWT", nil)

	c.
		On("GetUserInfo", "a JWT").
		Return(signin.UserInfoResponse{Email: "user@example.org"}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/auth/callback?code=auth-code", nil)

	SignInCallback(c, "/app/public-url", "client-id", "jti")(w, r)
	resp := w.Result()
	location, _ := resp.Location()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, location.String(), "/app/public-url/home?email=user%40example.org")
	mock.AssertExpectationsForObjects(t, c)
}
