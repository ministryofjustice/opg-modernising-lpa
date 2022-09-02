package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const formUrlEncoded = "application/x-www-form-urlencoded"

type mockDataStore struct {
	mock.Mock
}

func (m *mockDataStore) Get(ctx context.Context, id string, v interface{}) error {
	return m.Called(ctx, id, v).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, id string, v interface{}) error {
	return m.Called(ctx, id, v).Error(0)
}

func withSession(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), sessionKey{}, "session-id"))
}

func TestApp(t *testing.T) {
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{}, nil, nil)

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestLangRedirect(t *testing.T) {
	testCases := map[Lang]string{
		En: "/somewhere",
		Cy: "/cy/somewhere",
	}

	for lang, url := range testCases {
		t.Run("En", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			lang.Redirect(w, r, "/somewhere", http.StatusFound)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestFakeAddressClient(t *testing.T) {
	addresses, _ := fakeAddressClient{}.LookupPostcode("xyz")

	assert.Equal(t, []Address{
		{Line1: "123 Fake Street", TownOrCity: "Someville", Postcode: "xyz"},
		{Line1: "456 Fake Street", TownOrCity: "Someville", Postcode: "xyz"},
	}, addresses)
}

func TestRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"email": "person@example.com"}}, nil)

	handler := makeRequireSession(nil, sessionsStore)(http.NotFoundHandler())

	handler.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestRequireSessionNotAuthenticated(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := &mockLogger{}
	logger.
		On("Print", "email missing from session")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{}, nil)

	handler := makeRequireSession(logger, sessionsStore)(http.NotFoundHandler())

	handler.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, logger, sessionsStore)
}

func TestTestingStart(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere", nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{}, nil)
	sessionsStore.
		On("Save", r, w, &sessions.Session{Values: map[interface{}]interface{}{"email": "testing@example.com"}}).
		Return(nil)

	testingStart(sessionsStore).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestCookieConsentSet(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	assert.False(t, cookieConsentSet(r))

	r.AddCookie(&http.Cookie{Name: "cookies-consent"})
	assert.True(t, cookieConsentSet(r))
}
