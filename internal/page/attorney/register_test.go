package attorney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, nil, template.Templates{}, template.Templates{}, nil, nil, nil, nil, nil, nil, nil, nil, &mockDashboardStore{}, &lpastore.Client{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"session": &sesh.LoginSession{
					Sub: "random",
				},
			},
		}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			CanGoBack: false,
			SessionID: "cmFuZG9t",
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleRequireSessionExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			CanGoBack: true,
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

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, errorHandler.Execute)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleLpaMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
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

func TestMakeAttorneyHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "lpa-id", SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)
	expectedDetails := &actor.AttorneyProvidedDetails{ID: "123"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(expectedDetails, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore)
	handle("/path", CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, details *actor.AttorneyProvidedDetails) error {
		assert.Equal(t, expectedDetails, details)

		assert.Equal(t, page.AppData{
			Page:       "/attorney/lpa-id/path",
			CanGoBack:  true,
			SessionID:  "cmFuZG9t",
			LpaID:      "lpa-id",
			ActorType:  actor.TypeAttorney,
			AttorneyID: "123",
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())

		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t", LpaID: "lpa-id"}, sessionData)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeAttorneyHandleExistingLpaData(t *testing.T) {
	testCases := map[string]struct {
		details   *actor.AttorneyProvidedDetails
		actorType actor.Type
	}{
		"attorney": {
			details:   &actor.AttorneyProvidedDetails{ID: "attorney-id"},
			actorType: actor.TypeAttorney,
		},
		"replacement attorney": {
			details:   &actor.AttorneyProvidedDetails{ID: "attorney-id", IsReplacement: true},
			actorType: actor.TypeReplacementAttorney,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "lpa-id", SessionID: "ignored-session-id"})
			w := httptest.NewRecorder()
			r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(r, "session").
				Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Get(mock.Anything).
				Return(tc.details, nil)

			mux := http.NewServeMux()
			handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore)
			handle("/path", CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, details *actor.AttorneyProvidedDetails) error {
				assert.Equal(t, tc.details, details)
				assert.Equal(t, page.AppData{
					Page:       "/attorney/lpa-id/path",
					CanGoBack:  true,
					LpaID:      "lpa-id",
					SessionID:  "cmFuZG9t",
					AttorneyID: "attorney-id",
					ActorType:  tc.actorType,
				}, appData)
				assert.Equal(t, w, hw)

				sessionData, _ := page.SessionDataFromContext(hr.Context())

				assert.Equal(t, &page.SessionData{LpaID: "lpa-id", SessionID: "cmFuZG9t"}, sessionData)
				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}
}

func TestMakeAttorneyHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, errorHandler.Execute, attorneyStore)
	handle("/path", None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.AttorneyProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeAttorneyHandleAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, errorHandler.Execute, attorneyStore)
	handle("/path", None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.AttorneyProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
}

func TestMakeAttorneyHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, nil)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.AttorneyProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeAttorneyHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, nil)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.AttorneyProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeAttorneyHandleLpaMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{}, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, nil)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.AttorneyProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start.Format(), resp.Header.Get("Location"))
}
