package donor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &sesh.DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, None)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			CanGoBack: true,
			SessionID: "cmFuZG9t",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t"}, page.SessionDataFromContext(hr.Context()))
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &sesh.DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, None)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			SessionID: "cmFuZG9t",
			CanGoBack: true,
			LpaID:     "123",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, &page.SessionData{LpaID: "123", SessionID: "cmFuZG9t"}, page.SessionDataFromContext(hr.Context()))
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	logger := &page.MockLogger{}
	logger.
		On("Print", fmt.Sprintf("Error rendering page for path '%s': %s", "/path", page.ExpectedError.Error()))

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &sesh.DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return page.ExpectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	logger := &page.MockLogger{}
	logger.
		On("Print", page.ExpectedError)

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, page.ExpectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	logger := &page.MockLogger{}
	logger.
		On("Print", sesh.MissingSessionError("donor"))

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, None)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page: "/path",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(page.ContextWithAppData(r.Context(), page.AppData{Page: "/path"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestRouteToLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/somewhere%2Fwhat", nil)

	routeToLpa(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/somewhere/what", r.URL.Path)
		assert.Equal(t, "/somewhere%2Fwhat", r.URL.RawPath)

		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(w, r)

	res := w.Result()

	assert.Equal(t, http.StatusTeapot, res.StatusCode)
}

func TestRouteToLpaWithoutID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/", nil)

	routeToLpa(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(w, r)

	res := w.Result()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
