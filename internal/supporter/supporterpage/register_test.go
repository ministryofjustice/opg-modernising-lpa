package supporterpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &mockLogger{}, template.Templates{}, &onelogin.Client{}, &mockSessionStore{}, &mockOrganisationStore{}, nil, &notify.Client{}, "http://base", &mockMemberStore{}, &search.Client{}, &mockDonorStore{}, &mockShareCodeStore{}, &mockProgressTracker{}, &lpastore.ResolvingService{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:          "/path",
			SupporterData: &appcontext.SupporterData{},
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
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
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
	handle("/path", RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:          "/path",
			SessionID:     "cmFuZG9t",
			SupporterData: &appcontext.SupporterData{},
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
	assert.Equal(t, page.PathSupporterStart.Format(), resp.Header.Get("Location"))
}

func TestMakeSupporterHandle(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &appcontext.SupporterData{
				Permission:          supporterdata.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			CanGoBack:         true,
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())
		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWithLpaPath(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path/xyz", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org", LpaID: "xyz"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(supporter.LpaPath("/path"), CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/supporter/path/xyz",
			SessionID: "cmFuZG9t",
			SupporterData: &appcontext.SupporterData{
				Permission:          supporterdata.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			CanGoBack:         true,
			LoginSessionEmail: "a@example.org",
			LpaID:             "xyz",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())
		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id", LpaID: "xyz"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWhenRequireAdmin(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), RequireAdmin, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &appcontext.SupporterData{
				Permission:          supporterdata.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())
		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeSupporterHandleWhenRequireAdminAsNonAdmin(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionNone, ID: "member-id"}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, mock.Anything)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
}

func TestMakeSupporterHandleWhenSuspended(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id", Name: "My Org"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id", Status: supporterdata.StatusSuspended}, nil)

	suspendedTmpl := newMockTemplate(t)
	suspendedTmpl.EXPECT().
		Execute(w, &suspendedData{
			App: appcontext.Data{
				Page:              "/supporter/path",
				SessionID:         "cmFuZG9t",
				SupporterData:     &appcontext.SupporterData{},
				LoginSessionEmail: "a@example.org",
			},
			OrganisationName: "My Org",
		}).
		Return(nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, suspendedTmpl.Execute)
	handle(supporter.Path("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeSupporterHandleWhenSuspendedTemplateErrors(t *testing.T) {
	ctx := appcontext.ContextWithData(context.Background(), appcontext.Data{})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/supporter/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", OrganisationID: "org-id", Email: "a@example.org"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Organisation{ID: "org-id", Name: "My Org"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id", Status: supporterdata.StatusSuspended}, nil)

	suspendedTmpl := newMockTemplate(t)
	suspendedTmpl.EXPECT().
		Execute(w, &suspendedData{
			App: appcontext.Data{
				Page:              "/supporter/path",
				SessionID:         "cmFuZG9t",
				SupporterData:     &appcontext.SupporterData{},
				LoginSessionEmail: "a@example.org",
			},
			OrganisationName: "My Org",
		}).
		Return(expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, suspendedTmpl.Execute)
	handle(supporter.Path("/path"), RequireAdmin, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeSupporterHandleWithSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		appcontext.ContextWithSession(context.Background(),
			&appcontext.Session{SessionID: "existing-sub", OrganisationID: "an-org-id"}),
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
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id"})).
		Return(&supporterdata.Organisation{ID: "org-id"}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", Email: "a@example.org"})).
		Return(&supporterdata.Member{Permission: supporterdata.PermissionAdmin, ID: "member-id"}, nil)

	handle := makeSupporterHandle(mux, sessionStore, nil, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/supporter/path",
			SessionID: "cmFuZG9t",
			SupporterData: &appcontext.SupporterData{
				Permission:          supporterdata.PermissionAdmin,
				LoggedInSupporterID: "member-id",
			},
			LoginSessionEmail: "a@example.org",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())
		assert.Equal(t, &appcontext.Session{SessionID: "cmFuZG9t", Email: "a@example.org", OrganisationID: "org-id"}, sessionData)

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
	handle(supporter.Path("/path"), None, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterStart.Format(), resp.Header.Get("Location"))
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
	handle(supporter.Path("/path"), None, nil)

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
		Return(&supporterdata.Organisation{}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(mock.Anything).
		Return(&supporterdata.Member{}, expectedError)

	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), None, nil)

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
		Return(&supporterdata.Organisation{}, nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Get(mock.Anything).
		Return(&supporterdata.Member{}, nil)

	mux := http.NewServeMux()
	handle := makeSupporterHandle(mux, sessionStore, errorHandler.Execute, organisationStore, memberStore, nil)
	handle(supporter.Path("/path"), None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *supporterdata.Organisation, _ *supporterdata.Member) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
