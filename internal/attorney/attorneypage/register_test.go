package attorneypage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &mockLogger{}, template.Templates{}, template.Templates{}, nil, nil, nil, nil, nil, &mockDashboardStore{}, &lpastore.Client{}, &lpastore.ResolvingService{}, &mockNotifyClient{}, "http://app")

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

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
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

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
		Login(r).
		Return(nil, expectedError)

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
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/attorney/lpa-id/path?a=b", nil)
	expectedDetails := &actor.AttorneyProvidedDetails{UID: uid}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(expectedDetails, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore)
	handle("/path", CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, details *actor.AttorneyProvidedDetails) error {
		assert.Equal(t, expectedDetails, details)

		assert.Equal(t, page.AppData{
			Page:        "/attorney/lpa-id/path",
			CanGoBack:   true,
			SessionID:   "cmFuZG9t",
			LpaID:       "lpa-id",
			ActorType:   actor.TypeAttorney,
			AttorneyUID: uid,
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
	uid := actoruid.New()
	testCases := map[string]struct {
		details   *actor.AttorneyProvidedDetails
		actorType actor.Type
	}{
		"attorney": {
			details:   &actor.AttorneyProvidedDetails{UID: uid},
			actorType: actor.TypeAttorney,
		},
		"replacement attorney": {
			details:   &actor.AttorneyProvidedDetails{UID: uid, IsReplacement: true},
			actorType: actor.TypeReplacementAttorney,
		},
		"trust corporation": {
			details:   &actor.AttorneyProvidedDetails{UID: uid, IsTrustCorporation: true},
			actorType: actor.TypeTrustCorporation,
		},
		"replacement trust corporation": {
			details:   &actor.AttorneyProvidedDetails{UID: uid, IsReplacement: true, IsTrustCorporation: true},
			actorType: actor.TypeReplacementTrustCorporation,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			w := httptest.NewRecorder()
			r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/attorney/lpa-id/path?a=b", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Sub: "random"}, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Get(mock.Anything).
				Return(tc.details, nil)

			mux := http.NewServeMux()
			handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore)
			handle("/path", CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, details *actor.AttorneyProvidedDetails) error {
				assert.Equal(t, tc.details, details)
				assert.Equal(t, page.AppData{
					Page:        "/attorney/lpa-id/path",
					CanGoBack:   true,
					LpaID:       "lpa-id",
					SessionID:   "cmFuZG9t",
					AttorneyUID: uid,
					ActorType:   tc.actorType,
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

func TestMakeAttorneyHandleExistingSessionDataWhenCannotGoToURL(t *testing.T) {
	path := page.Paths.Attorney.Sign

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, path.Format("123"), nil)
	expectedDetails := &actor.AttorneyProvidedDetails{UID: uid, LpaID: "123"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(expectedDetails, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore)
	handle(path, CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, details *actor.AttorneyProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.TaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeAttorneyHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/attorney/id/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

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
	r, _ := http.NewRequest(http.MethodGet, "/attorney/id/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

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
	r, _ := http.NewRequest(http.MethodGet, "/attorney/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

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
