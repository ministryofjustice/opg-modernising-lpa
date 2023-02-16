package certificateprovider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &log.Logger{}, template.Templates{}, nil, nil, &onelogin.Client{}, nil, &place.Client{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"certificate-provider": &sesh.CertificateProviderSession{
					Sub:            "random",
					DonorSessionID: "session-id",
					LpaID:          "lpa-id",
				},
			},
		}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			SessionID: "session-id",
			LpaID:     "lpa-id",
			CanGoBack: false,
		}, appData)
		assert.Equal(t, w, hw)

		assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, page.SessionDataFromContext(hr.Context()))
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "ignored-123", SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"certificate-provider": &sesh.CertificateProviderSession{Sub: "random", LpaID: "lpa-id", DonorSessionID: "session-id"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, None)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			SessionID: "session-id",
			CanGoBack: true,
			LpaID:     "lpa-id",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, &page.SessionData{LpaID: "lpa-id", SessionID: "session-id"}, page.SessionDataFromContext(hr.Context()))
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

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("Error rendering page for path '%s': %s", "/path", expectedError.Error()))

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, nil, None)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderStart, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	logger := &mockLogger{}
	logger.
		On("Print", sesh.MissingSessionError("certificate-provider"))

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, None)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderStart, resp.Header.Get("Location"))
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
