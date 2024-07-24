package certificateprovider

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
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
	Register(mux, &slog.Logger{}, template.Templates{}, template.Templates{}, nil, &onelogin.Client{}, nil, nil, nil, &place.Client{}, &notify.Client{}, nil, &mockDashboardStore{}, &lpastore.Client{}, &lpastore.ResolvingService{}, &mockDonorStore{}, "publicURL")

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil)
	handle("/path", func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeCertificateProvider,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())

		assert.Nil(t, sessionData)
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
	handle := makeHandle(mux, errorHandler.Execute)
	handle("/path", func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeCertificateProviderHandle(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/certificate-provider/123/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, certificateProviderStore)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, certificateProvider *actor.CertificateProviderProvidedDetails) error {
		assert.Equal(t, page.AppData{
			Page:      "/certificate-provider/123/path",
			SessionID: "cmFuZG9t",
			LpaID:     "123",
			ActorType: actor.TypeCertificateProvider,
		}, appData)
		assert.Equal(t, w, hw)

		assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)

		sessionData, _ := page.SessionDataFromContext(hr.Context())
		assert.Equal(t, &page.SessionData{LpaID: "123", SessionID: "cmFuZG9t"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeCertificateProviderHandleWhenCannotGoToURL(t *testing.T) {
	path := page.Paths.CertificateProvider.ProvideCertificate

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, path.Format("123"), nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, certificateProviderStore)
	handle(path, page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.CertificateProviderProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.TaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeCertificateProviderHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/certificate-provider/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, nil, nil)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.CertificateProviderProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderStart.Format(), resp.Header.Get("Location"))
}

func TestMakeCertificateProviderHandleWhenAttorneyStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored-session-id"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/certificate-provider/id/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeCertificateProviderHandle(mux, sessionStore, errorHandler.Execute, certificateProviderStore)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.CertificateProviderProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
}
