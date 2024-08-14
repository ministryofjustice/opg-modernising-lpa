package voucherpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &mockLogger{}, template.Templates{}, &mockSessionStore{}, &mockVoucherStore{}, &mockOneLoginClient{}, &mockShareCodeStore{}, &mockDashboardStore{}, nil, &mockLpaStoreResolvingService{})

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
	handle("/path", RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			CanGoBack: false,
			ActorType: actor.TypeVoucher,
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
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession|CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			CanGoBack: true,
			ActorType: actor.TypeVoucher,
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
	handle := makeHandle(mux, nil, errorHandler.Execute)
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
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathVoucherStart.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			ActorType: actor.TypeVoucher,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(appcontext.ContextWithData(r.Context(), appcontext.Data{Page: "/path", ActorType: actor.TypeVoucher})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeVoucherHandleExistingSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/voucher/lpa-id/task-list?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Get(mock.Anything).
		Return(&voucherdata.Provided{LpaID: "lpa-id"}, nil)

	mux := http.NewServeMux()
	handle := makeVoucherHandle(mux, sessionStore, nil, voucherStore)
	handle(voucher.PathTaskList, CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, provided *voucherdata.Provided) error {
		assert.Equal(t, &voucherdata.Provided{LpaID: "lpa-id"}, provided)

		assert.Equal(t, appcontext.Data{
			Page:      "/voucher/lpa-id/task-list",
			CanGoBack: true,
			ActorType: actor.TypeVoucher,
			SessionID: "cmFuZG9t",
			LpaID:     "lpa-id",
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

func TestMakeVoucherHandleWhenCannotGoToURL(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, voucher.PathSignTheDeclaration.Format("lpa-id"), nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Get(mock.Anything).
		Return(&voucherdata.Provided{LpaID: "lpa-id"}, nil)

	mux := http.NewServeMux()
	handle := makeVoucherHandle(mux, sessionStore, nil, voucherStore)
	handle(voucher.PathSignTheDeclaration, CanGoBack, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestMakeVoucherHandleWhenVoucherStoreErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/voucher/lpa-id/task-list?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeVoucherHandle(mux, sessionStore, errorHandler.Execute, voucherStore)
	handle(voucher.PathTaskList, CanGoBack, nil)

	mux.ServeHTTP(w, r)
}

func TestMakeVoucherHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/voucher/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Get(mock.Anything).
		Return(&voucherdata.Provided{LpaID: "lpa-id"}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeVoucherHandle(mux, sessionStore, errorHandler.Execute, voucherStore)
	handle("/path", None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *voucherdata.Provided) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeVoucherHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/voucher/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeVoucherHandle(mux, sessionStore, nil, nil)
	handle("/path", RequireSession, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *voucherdata.Provided) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathVoucherStart.Format(), resp.Header.Get("Location"))
}
