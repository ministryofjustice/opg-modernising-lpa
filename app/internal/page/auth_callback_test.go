package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthCallbackClient struct {
	mock.Mock
}

func (m *mockAuthCallbackClient) Exchange(code string) (string, error) {
	args := m.Called(code)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockAuthCallbackClient) UserInfo(jwt string) (signin.UserInfo, error) {
	args := m.Called(jwt)
	return args.Get(0).(signin.UserInfo), args.Error(1)
}

func TestSignInCallback(t *testing.T) {
	w := httptest.NewRecorder()

	client := &mockAuthCallbackClient{}
	client.
		On("Exchange", "auth-code").
		Return("a JWT", nil)
	client.
		On("UserInfo", "a JWT").
		Return(signin.UserInfo{Email: "user@example.org"}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/auth/callback?code=auth-code", nil)

	AuthCallback(nil, client)(w, r)
	resp := w.Result()
	location, _ := resp.Location()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, location.String(), "/home?email=user%40example.org")
	mock.AssertExpectationsForObjects(t, client)
}

func TestSignInCallbackWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockAuthCallbackClient{}
	client.
		On("Exchange", "auth-code").
		Return("", expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	AuthCallback(logger, client)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}

func TestSignInCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockAuthCallbackClient{}
	client.
		On("Exchange", "auth-code").
		Return("a JWT", nil)
	client.
		On("UserInfo", "a JWT").
		Return(signin.UserInfo{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	AuthCallback(logger, client)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}
