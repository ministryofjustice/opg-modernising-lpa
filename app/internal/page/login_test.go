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

func (m *mockLoginClient) AuthCodeURL(state, nonce, locale string) string {
	args := m.Called(state, nonce, locale)

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
	r, _ := http.NewRequest(http.MethodGet, "/?locale=blah", nil)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random", "blah").
		Return("http://auth?locale=blah")

	sessionsStore := &mockSessionsStore{}

	session := sessions.NewSession(sessionsStore, "params")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[interface{}]interface{}{"state": "i am random", "nonce": "i am random"}

	sessionsStore.
		On("Save", r, w, session).
		Return(nil)

	Login(nil, client, sessionsStore, true, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth?locale=blah", resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, client, sessionsStore)
}

func TestLoginDefaultLocale(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random", "en").
		Return("http://auth?locale=en")

	sessionsStore := &mockSessionsStore{}

	session := sessions.NewSession(sessionsStore, "params")

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[interface{}]interface{}{"state": "i am random", "nonce": "i am random"}

	sessionsStore.
		On("Save", r, w, session).
		Return(nil)

	Login(nil, client, sessionsStore, true, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://auth?locale=en", resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, client, sessionsStore)
}

func TestLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	client := &mockLoginClient{}
	client.
		On("AuthCodeURL", "i am random", "i am random", "en").
		Return("http://auth?locale=en")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	Login(logger, client, sessionsStore, true, func(int) string { return "i am random" })(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, logger, client, sessionsStore)
}
