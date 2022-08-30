package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
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
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hey", nil)

	client := &mockAuthRedirectClient{}
	client.
		On("Exchange", "auth-code").
		Return("a JWT", nil)
	client.
		On("UserInfo", "a JWT").
		Return(signin.UserInfo{Email: "user@example.org"}, nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"state": "hey"}}, nil)

	AuthRedirect(nil, client, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, resp.Header.Get("Location"), "/home?email=user%40example.org")
	mock.AssertExpectationsForObjects(t, client, sessionsStore)
}

func TestAuthRedirectStateMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code", nil)

	logger := &mockLogger{}
	logger.
		On("Print", "state missing or incorrect")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"state": "hey"}}, nil)

	AuthRedirect(logger, nil, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, logger, sessionsStore)
}

func TestAuthRedirectStateIncorrect(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hello", nil)

	logger := &mockLogger{}
	logger.
		On("Print", "state missing or incorrect")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"state": "hey"}}, nil)

	AuthRedirect(logger, nil, sessionsStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, logger, sessionsStore)
}

func TestAuthRedirectWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hey", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockAuthRedirectClient{}
	client.
		On("Exchange", "auth-code").
		Return("", expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"state": "hey"}}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}

func TestAuthRedirectWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=hey", nil)

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

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"state": "hey"}}, nil)

	AuthRedirect(logger, client, sessionsStore)(w, r)

	mock.AssertExpectationsForObjects(t, client, logger)
}
