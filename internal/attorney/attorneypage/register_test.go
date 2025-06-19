package attorneypage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &mockLogger{}, template.Templates{}, template.Templates{}, nil, nil, nil, nil, nil, &mockDashboardStore{}, &lpastore.Client{}, &lpastore.ResolvingService{}, &mockNotifyClient{}, &mockEventClient{}, &mockBundle{}, "attorneyStartURL")

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
	handle := makeHandle(mux, sessionStore, nil, "")
	handle("/path", RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
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

func TestMakeHandleRequireSessionExistingSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil, "")
	handle("/path", RequireSession|CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			CanGoBack: true,
			SessionID: "cmFuZG9t",
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())

		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t"}, sessionData)
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
	handle := makeHandle(mux, nil, errorHandler.Execute, "")
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
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
	handle := makeHandle(mux, sessionStore, nil, "http://example.com")
	handle("/path", RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://example.com", resp.Header.Get("Location"))
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, "")
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page: "/path",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(appcontext.ContextWithData(r.Context(), appcontext.Data{Page: "/path"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeAttorneyHandleExistingSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/attorney/lpa-id/path?a=b", nil)

	expectedDetails := &attorneydata.Provided{UID: uid}
	expectedLpa := &lpadata.Lpa{LpaID: "lpa-id"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(expectedDetails, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(expectedLpa, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore, lpaStoreResolvingService, "")
	handle("/path", CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, details *attorneydata.Provided, lpa *lpadata.Lpa) error {
		assert.Equal(t, expectedDetails, details)
		assert.Equal(t, expectedLpa, lpa)

		assert.Equal(t, appcontext.Data{
			Page:        "/attorney/lpa-id/path",
			CanGoBack:   true,
			SessionID:   "cmFuZG9t",
			LpaID:       "lpa-id",
			ActorType:   actor.TypeAttorney,
			AttorneyUID: uid,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())

		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t", LpaID: "lpa-id"}, sessionData)
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
		details   *attorneydata.Provided
		actorType actor.Type
	}{
		"attorney": {
			details:   &attorneydata.Provided{UID: uid},
			actorType: actor.TypeAttorney,
		},
		"replacement attorney": {
			details:   &attorneydata.Provided{UID: uid, IsReplacement: true},
			actorType: actor.TypeReplacementAttorney,
		},
		"trust corporation": {
			details:   &attorneydata.Provided{UID: uid, IsTrustCorporation: true},
			actorType: actor.TypeTrustCorporation,
		},
		"replacement trust corporation": {
			details:   &attorneydata.Provided{UID: uid, IsReplacement: true, IsTrustCorporation: true},
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

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(mock.Anything).
				Return(&lpadata.Lpa{}, nil)

			mux := http.NewServeMux()
			handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore, lpaStoreResolvingService, "")
			handle("/path", CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, details *attorneydata.Provided, lpa *lpadata.Lpa) error {
				assert.Equal(t, tc.details, details)
				assert.Equal(t, appcontext.Data{
					Page:        "/attorney/lpa-id/path",
					CanGoBack:   true,
					LpaID:       "lpa-id",
					SessionID:   "cmFuZG9t",
					AttorneyUID: uid,
					ActorType:   tc.actorType,
				}, appData)
				assert.Equal(t, w, hw)

				sessionData, _ := appcontext.SessionFromContext(hr.Context())

				assert.Equal(t, &appcontext.Session{LpaID: "lpa-id", SessionID: "cmFuZG9t"}, sessionData)
				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}
}

func TestMakeAttorneyHandleExistingSessionWhenCannotGoToURL(t *testing.T) {
	path := attorney.PathSign

	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, path.Format("123"), nil)
	expectedDetails := &attorneydata.Provided{UID: uid, LpaID: "123"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(expectedDetails, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, nil, attorneyStore, lpaStoreResolvingService, "")
	handle(path, CanGoBack, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathTaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeAttorneyHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/attorney/id/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(&attorneydata.Provided{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, errorHandler.Execute, attorneyStore, lpaStoreResolvingService, "")
	handle("/path", None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *attorneydata.Provided, _ *lpadata.Lpa) error {
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
	handle := makeAttorneyHandle(mux, sessionStore, errorHandler.Execute, attorneyStore, nil, "")
	handle("/path", None, nil)

	mux.ServeHTTP(w, r)
}

func TestMakeAttorneyHandleLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/attorney/id/path", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Get(mock.Anything).
		Return(&attorneydata.Provided{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeAttorneyHandle(mux, sessionStore, errorHandler.Execute, attorneyStore, lpaStoreResolvingService, "")
	handle("/path", None, nil)

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
	handle := makeAttorneyHandle(mux, sessionStore, nil, nil, nil, "http://example.com")
	handle("/path", RequireSession, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://example.com", resp.Header.Get("Location"))
}
