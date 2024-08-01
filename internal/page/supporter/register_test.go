package supporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &mockLogger{}, template.Templates{}, &onelogin.Client{}, &mockSessionStore{}, &mockOrganisationStore{}, nil, &notify.Client{}, "http://base", &mockMemberStore{}, &search.Client{}, &mockDonorStore{}, &mockShareCodeStore{}, &mockProgressTracker{}, &lpastore.ResolvingService{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:          "/path",
			SupporterData: &page.SupporterData{},
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
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeHandleWhenRequireSession(t *testing.T) {
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
			Page:          "/path",
			SessionID:     "cmFuZG9t",
			SupporterData: &page.SupporterData{},
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
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeSupporterHandle(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		assert.Equal(t, page.AppData{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &page.SupporterData{
				Permission:          actor.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			CanGoBack:         true,
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionDataFromContext(hr.Context())
		assert.Equal(t, &appcontext.SessionData{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWithLpaPath(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path/xyz", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org", LpaID: "xyz"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(page.SupporterLpaPath("/path"), CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		assert.Equal(t, page.AppData{
			Page:      "/supporter/path/xyz",
			SessionID: "cmFuZG9t",
			SupporterData: &page.SupporterData{
				Permission:          actor.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			CanGoBack:         true,
			LoginSessionEmail: "a@example.org",
			LpaID:             "xyz",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionDataFromContext(hr.Context())
		assert.Equal(t, &appcontext.SessionData{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id", LpaID: "xyz"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWhenRequireAdmin(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), RequireAdmin, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		assert.Equal(t, page.AppData{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &page.SupporterData{
				Permission:          actor.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionDataFromContext(hr.Context())
		assert.Equal(t, &appcontext.SessionData{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWhenRequireAdminAsNonAdmin(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionNone, ID: "member-id"}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, mock.Anything)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
}

func TestMakeSupporterHandleWhenSuspended(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id", Name: "My Org"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id", Status: actor.StatusSuspended}, nil)

	suspendedTmpl := newMockTemplate(t)
	suspendedTmpl.EXPECT().
		Execute(w, &suspendedData{
			App: page.AppData{
				Page:              "/supporter/path",
				SessionID:         "cmFuZG9t",
				SupporterData:     &page.SupporterData{},
				LoginSessionEmail: "a@example.org",
			},
			OrganisationName: "My Org",
		}).
		Return(nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, suspendedTmpl.Execute)
	handle(page.SupporterPath("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeSupporterHandleWhenSuspendedTemplateErrors(t *testing.T) {
	ctx := page.ContextWithAppData(context.Background(), page.AppData{CanToggleWelsh: true})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Organisation{ID: "org-id", Name: "My Org"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id", Status: actor.StatusSuspended}, nil)

	suspendedTmpl := newMockTemplate(t)
	suspendedTmpl.EXPECT().
		Execute(w, &suspendedData{
			App: page.AppData{
				Page:              "/supporter/path",
				SessionID:         "cmFuZG9t",
				SupporterData:     &page.SupporterData{},
				LoginSessionEmail: "a@example.org",
			},
			OrganisationName: "My Org",
		}).
		Return(expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, suspendedTmpl.Execute)
	handle(page.SupporterPath("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeSupporterHandleWithSessionData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		page.ContextWithSessionData(context.Background(),
			&appcontext.SessionData{SessionID: "existing-sub", OrganisationID: "an-org-id"}),
		http.MethodGet,
		"/supporter/path",
		nil,
	)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id"})).
		Return(&actor.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&actor.Member{Permission: actor.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		assert.Equal(t, page.AppData{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &page.SupporterData{
				Permission:          actor.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionDataFromContext(hr.Context())
		assert.Equal(t, &appcontext.SessionData{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

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
		Login(r).
		Return(nil, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, nil, nil, nil, nil)
	handle(page.SupporterPath("/path"), None, nil)

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
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, nil, nil)
	handle(page.SupporterPath("/path"), None, nil)

	mux.ServeHTTP(w, r)
}

func TestMakeSupporterHandleWhenMemberStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Member{}, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), None, nil)

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
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Member{}, nil)

	mux := http.NewServeMux()
	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(page.SupporterPath("/path"), None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.Organisation, _ *actor.Member) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
