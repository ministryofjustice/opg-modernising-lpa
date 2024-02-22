package supporter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")
var testAppData = page.AppData{}

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, template.Templates{}, &onelogin.Client{}, nil, nil, nil, &notify.Client{}, "http://base", nil, &search.Client{}, nil)

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:        "/path",
			IsSupporter: true,
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, errorHandler.Execute)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeHandleWhenRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:        "/path",
			SessionID:   "cmFuZG9t",
			IsSupporter: true,
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleWhenRequireSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeSupporterHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}}}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore)
	handle("/path", page.CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation) error {
		assert.Equal(t, page.AppData{
			Page:              "/supporter/path",
			SessionID:         "cmFuZG9t",
			IsSupporter:       true,
			CanGoBack:         true,
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())
		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t", Email: "a@example.org"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWithSessionData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		page.ContextWithSessionData(context.Background(),
			&page.SessionData{SessionID: "existing-sub", OrganisationID: "an-org-id"}),
		http.MethodGet,
		"/supporter/path",
		nil,
	)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random", OrganisationID: "org-id"}}}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id"})).
		Return(&actor.Organisation{}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation) error {
		assert.Equal(t, page.AppData{
			Page:        "/supporter/path",
			SessionID:   "cmFuZG9t",
			IsSupporter: true,
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

func TestMakeSupporterHandleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{}, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, nil, nil)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.Organisation) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeSupporterHandleWhenOrganisationStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation) error {
		return nil
	})

	mux.ServeHTTP(w, r)
}

func TestMakeSupporterHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/supporter/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	mux := http.NewServeMux()
	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.Organisation) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
