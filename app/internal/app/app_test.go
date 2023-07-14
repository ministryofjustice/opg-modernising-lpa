package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/uid"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App(&logging.Logger{}, &localize.Localizer{}, localize.En, template.Templates{}, nil, &dynamo.Client{}, "http://public.url", &pay.Client{}, &identity.YotiClient{}, &notify.Client{}, &place.Client{}, page.RumConfig{}, "?%3fNEI0t9MN", page.Paths, &onelogin.Client{}, &uid.Client{}, "http://onelogin.url", &s3.Client{}, "bucket")

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestQueryString(t *testing.T) {
	testCases := map[string]struct {
		url           string
		expectedQuery string
	}{
		"with query": {
			url:           "http://example.org/?a=query&b=string",
			expectedQuery: "?a=query&b=string",
		},
		"with empty query": {
			url:           "http://example.org/?",
			expectedQuery: "",
		},
		"without query": {
			url:           "http://example.org/",
			expectedQuery: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			assert.Equal(t, tc.expectedQuery, queryString(r))
		})
	}
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page: "/path",
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			SessionID: "cmFuZG9t",
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())
		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleRequireSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleRequireSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, errorHandler.Execute, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
