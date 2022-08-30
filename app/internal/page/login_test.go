package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLoginClient struct {
	mock.Mock
}

func (m *mockLoginClient) AuthCodeURL(state, nonce string) string {
	args := m.Called(state, nonce)

	return args.String(0)
}

type mockSessionsStore struct {
	mock.Mock
}

func (m *mockSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	args := m.Called(r, w, session)
	return args.Error(0)
}

func TestLogin(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random").
		Return("http://auth")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("New", r, "params").
		Return(&sessions.Session{}, nil)
	sessionsStore.
		On("Save", r, w, &sessions.Session{Values: map[interface{}]interface{}{"state": "i am random", "nonce": "i am random"}}).
		Return(nil)

	Login(nil, client, sessionsStore, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth", resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, client, sessionsStore)
}

func TestLoginWhenStoreNewError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random").
		Return("http://auth")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("New", r, "params").
		Return(&sessions.Session{}, expectedError)

	Login(logger, client, sessionsStore, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, logger, client, sessionsStore)
}

func TestLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random").
		Return("http://auth")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("New", r, "params").
		Return(&sessions.Session{}, nil)
	sessionsStore.
		On("Save", r, w, &sessions.Session{Values: map[interface{}]interface{}{"state": "i am random", "nonce": "i am random"}}).
		Return(expectedError)

	Login(logger, client, sessionsStore, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, logger, client, sessionsStore)
}
