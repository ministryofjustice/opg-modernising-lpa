package page

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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

var (
	expectedError = errors.New("err")
	appData       = AppData{SessionID: "session-id"}
)

type mockDataStore struct {
	data interface{}
	mock.Mock
}

func (m *mockDataStore) Get(ctx context.Context, id string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, id).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, id string, v interface{}) error {
	return m.Called(ctx, id, v).Error(0)
}

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
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

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"email": "person@example.com"}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, En)
	handle("/path", true, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Localizer:        localizer,
			Lang:             En,
			SessionID:        "cGVyc29uQGV4YW1wbGUuY29t",
			CookieConsentSet: false,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r, hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, En)
	handle("/path", true, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", "email missing from session")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, En)
	handle("/path", true, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, localizer, En)
	handle("/path", false, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Localizer:        localizer,
			Lang:             En,
			CookieConsentSet: false,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r, hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestTestingStart(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere", nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "params").
		Return(&sessions.Session{}, nil)
	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(nil)

	testingStart(sessionsStore).ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore)
}
