package certificateproviderpage

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &slog.Logger{}, template.Templates{}, template.Templates{}, nil, &onelogin.Client{}, nil, nil, nil, &place.Client{}, &notify.Client{}, nil, &mockDashboardStore{}, &lpastore.Client{}, &lpastore.ResolvingService{}, &mockDonorStore{}, &mockEventClient{}, &mockScheduledStore{}, "publicURL", "donorStartURL")

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			ActorType: actor.TypeCertificateProvider,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())

		assert.Nil(t, sessionData)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleWhenSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", page.RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/path",
			ActorType: actor.TypeCertificateProvider,
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

func TestMakeHandleWhenSessionRequiredAndSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", page.RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathCertificateProviderStart.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, errorHandler.Execute, nil)
	handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeCertificateProviderHandle(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/certificate-provider/123/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&certificateproviderdata.Provided{LpaID: "123"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{LpaID: "123"}, nil)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, certificateProviderStore, lpaStoreResolvingService)
	handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/certificate-provider/123/path",
			SessionID: "cmFuZG9t",
			LpaID:     "123",
			ActorType: actor.TypeCertificateProvider,
		}, appData)
		assert.Equal(t, w, hw)

		assert.Equal(t, &certificateproviderdata.Provided{LpaID: "123"}, certificateProvider)
		assert.Equal(t, &lpadata.Lpa{LpaID: "123"}, lpa)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())
		assert.Equal(t, &appcontext.Session{LpaID: "123", SessionID: "cmFuZG9t"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeCertificateProviderHandleWhenCannotGoToURL(t *testing.T) {
	path := certificateprovider.PathProvideCertificate

	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, path.Format("123"), nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&certificateproviderdata.Provided{LpaID: "123"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, certificateProviderStore, lpaStoreResolvingService)
	handle(path, page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathTaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeCertificateProviderHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/certificate-provider/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, nil, nil)
	handle("/path", page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathCertificateProviderStart.Format(), resp.Header.Get("Location"))
}

func TestMakeCertificateProviderHandleWhenCertificateProviderStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/certificate-provider/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&certificateproviderdata.Provided{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, errorHandler.Execute, certificateProviderStore, nil)
	handle("/path", page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		return nil
	})

	mux.ServeHTTP(w, r)
}

func TestMakeCertificateProviderHandleWhenLpaStoreResolvingServiceError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/certificate-provider/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&certificateproviderdata.Provided{}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, errorHandler.Execute, certificateProviderStore, lpaStoreResolvingService)
	handle("/path", page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		return nil
	})

	mux.ServeHTTP(w, r)
}
