package donor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	lpastore "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &slog.Logger{}, template.Templates{}, template.Templates{}, &mockSessionStore{}, &mockDonorStore{}, &onelogin.Client{}, &place.Client{}, "http://example.org", &pay.Client{}, &mockShareCodeSender{}, &mockWitnessCodeSender{}, nil, &mockCertificateProviderStore{}, &mockNotifyClient{}, &mockEvidenceReceivedStore{}, &mockDocumentStore{}, &mockEventClient{}, &mockDashboardStore{}, &mockLpaStoreClient{}, &mockShareCodeStore{}, &mockProgressTracker{}, &lpastore.ResolvingService{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleWhenRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	loginSession := &sesh.LoginSession{Sub: "5", Email: "email"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(loginSession, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:              "/path",
			ActorType:         actor.TypeDonor,
			SessionID:         loginSession.SessionID(),
			LoginSessionEmail: "email",
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
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
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

func TestMakeLpaHandleWhenDetailsProvidedAndUIDExists(t *testing.T) {
	testCases := map[string]struct {
		expectedAppData     page.AppData
		loginSesh           *sesh.LoginSession
		expectedSessionData *page.SessionData
	}{
		"donor": {
			expectedAppData: page.AppData{
				Page:      "/lpa/123/path",
				ActorType: actor.TypeDonor,
				SessionID: "cmFuZG9t",
				LpaID:     "123",
			},
			loginSesh:           &sesh.LoginSession{Sub: "random"},
			expectedSessionData: &page.SessionData{SessionID: "cmFuZG9t", LpaID: "123"},
		},
		"organisation": {
			expectedAppData: page.AppData{
				Page:      "/lpa/123/path",
				ActorType: actor.TypeDonor,
				SessionID: "cmFuZG9t",
				LpaID:     "123",
				SupporterData: &page.SupporterData{
					DonorFullName: "Jane Smith",
					LpaType:       actor.LpaTypePropertyAndAffairs,
				},
			},
			loginSesh:           &sesh.LoginSession{Sub: "random", OrganisationID: "org-id"},
			expectedSessionData: &page.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", LpaID: "123"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/lpa/123/path", nil)

			mux := http.NewServeMux()

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(tc.loginSesh, nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(mock.Anything).
				Return(&actor.DonorProvidedDetails{Donor: actor.Donor{
					FirstNames:  "Jane",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Address:     place.Address{Postcode: "ABC123"},
				},
					Type:   actor.LpaTypePropertyAndAffairs,
					Tasks:  actor.DonorTasks{YourDetails: actor.TaskCompleted},
					LpaUID: "a-uid",
				}, nil)

			handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
			handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, _ *actor.DonorProvidedDetails) error {
				assert.Equal(t, tc.expectedAppData, appData)

				assert.Equal(t, w, hw)

				sessionData, _ := page.SessionDataFromContext(hr.Context())
				assert.Equal(t, tc.expectedSessionData, sessionData)

				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}

}

func TestMakeLpaHandleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/id/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	handle := makeLpaHandle(mux, sessionStore, nil, nil)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeLpaHandleWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/id/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeLpaHandleWhenCannotGoToURL(t *testing.T) {
	path := page.Paths.WhenCanTheLpaBeUsed
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, path.Format("123"), nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{LpaID: "123"}, nil)

	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle(path, page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeLpaHandleSessionExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/lpa/123/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle("/path", page.RequireSession|page.CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, _ *actor.DonorProvidedDetails) error {
		assert.Equal(t, page.AppData{
			Page:      "/lpa/123/path",
			SessionID: "cmFuZG9t",
			CanGoBack: true,
			LpaID:     "123",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())

		assert.Equal(t, &page.SessionData{LpaID: "123", SessionID: "cmFuZG9t"}, sessionData)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeLpaHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore)
	handle("/path", page.RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
