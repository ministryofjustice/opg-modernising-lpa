package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthRedirectClient struct {
	mock.Mock
}

func (m *mockAuthRedirectClient) Exchange(code string) (string, error) {
	args := m.Called(code)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockAuthRedirectClient) UserInfo(jwt string) (signin.UserInfo, error) {
	args := m.Called(jwt)
	return args.Get(0).(signin.UserInfo), args.Error(1)
}

func TestAuthRedirect(t *testing.T) {
	w := httptest.NewRecorder()

	client := &mockAuthRedirectClient{}
	client.
		On("Exchange", "auth-code").
		Return("a JWT", nil)
	client.
		On("UserInfo", "a JWT").
		Return(signin.UserInfo{Email: "user@example.org"}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	AuthRedirect(nil, client)(w, r)
	resp := w.Result()
	location, _ := resp.Location()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, location.String(), "/home?email=user%40example.org")
	mock.AssertExpectationsForObjects(t, client)
}

func TestAuthRedirectWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockAuthRedirectClient{}
	client.
		On("Exchange", "auth-code").
		Return("", expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	AuthRedirect(logger, client)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}

func TestAuthRedirectWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockAuthRedirectClient{}
	client.
		On("Exchange", "auth-code").
		Return("a JWT", nil)
	client.
		On("UserInfo", "a JWT").
		Return(signin.UserInfo{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	AuthRedirect(logger, client)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}
