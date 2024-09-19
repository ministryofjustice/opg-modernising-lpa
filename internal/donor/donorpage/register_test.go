package donorpage

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &slog.Logger{}, template.Templates{}, &mockSessionStore{}, &mockDonorStore{}, &onelogin.Client{}, &place.Client{}, "http://example.org", &pay.Client{}, &mockShareCodeSender{}, &mockWitnessCodeSender{}, nil, &mockCertificateProviderStore{}, &mockNotifyClient{}, &mockEvidenceReceivedStore{}, &mockDocumentStore{}, &mockEventClient{}, &mockDashboardStore{}, &mockLpaStoreClient{}, &mockShareCodeStore{}, &mockProgressTracker{}, &lpastore.ResolvingService{}, &mockScheduledStore{}, true)

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
	handle("/path", page.RequireSession, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, appcontext.Data{
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
	assert.Equal(t, page.PathStart.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, errorHandler.Execute)
	handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeLpaHandleWhenDetailsProvidedAndUIDExists(t *testing.T) {
	testCases := map[string]struct {
		expectedAppData appcontext.Data
		loginSesh       *sesh.LoginSession
		expectedSession *appcontext.Session
	}{
		"donor": {
			expectedAppData: appcontext.Data{
				Page:      "/lpa/123/path",
				ActorType: actor.TypeDonor,
				SessionID: "cmFuZG9t",
				LpaID:     "123",
			},
			loginSesh:       &sesh.LoginSession{Sub: "random"},
			expectedSession: &appcontext.Session{SessionID: "cmFuZG9t", LpaID: "123"},
		},
		"organisation": {
			expectedAppData: appcontext.Data{
				Page:      "/lpa/123/path",
				ActorType: actor.TypeDonor,
				SessionID: "cmFuZG9t",
				LpaID:     "123",
				SupporterData: &appcontext.SupporterData{
					DonorFullName: "Jane Smith",
					LpaType:       lpadata.LpaTypePropertyAndAffairs,
				},
			},
			loginSesh:       &sesh.LoginSession{Sub: "random", OrganisationID: "org-id"},
			expectedSession: &appcontext.Session{SessionID: "cmFuZG9t", OrganisationID: "org-id", LpaID: "123"},
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
				Return(&donordata.Provided{Donor: donordata.Donor{
					FirstNames:  "Jane",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Address:     place.Address{Postcode: "ABC123"},
					Email:       "a@example.com",
				},
					Type:   lpadata.LpaTypePropertyAndAffairs,
					Tasks:  donordata.Tasks{YourDetails: task.StateCompleted},
					LpaUID: "a-uid",
				}, nil)

			handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
			handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, _ *donordata.Provided) error {
				assert.Equal(t, tc.expectedAppData, appData)

				assert.Equal(t, w, hw)

				sessionData, _ := appcontext.SessionFromContext(hr.Context())
				assert.Equal(t, tc.expectedSession, sessionData)

				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}
}

func TestMakeHandleLpaWhenDonorEmailNotSet(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random", Email: "a@example.com"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&donordata.Provided{Donor: donordata.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:   lpadata.LpaTypePropertyAndAffairs,
			Tasks:  donordata.Tasks{YourDetails: task.StateCompleted},
			LpaUID: "a-uid",
		}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, &donordata.Provided{Donor: donordata.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
			Email:       "a@example.com",
		},
			Type:   lpadata.LpaTypePropertyAndAffairs,
			Tasks:  donordata.Tasks{YourDetails: task.StateCompleted},
			LpaUID: "a-uid",
		}).
		Return(nil)

	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle("/path", page.None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, _ *donordata.Provided) error {
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
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
	handle("/path", page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *donordata.Provided) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathStart.Format(), resp.Header.Get("Location"))
}

func TestMakeLpaHandleWhenDonorStoreError(t *testing.T) {
	testcases := map[string]func() *mockDonorStore{
		"get": func() *mockDonorStore {
			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(mock.Anything).
				Return(&donordata.Provided{}, expectedError)
			return donorStore
		},
		"put": func() *mockDonorStore {
			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(mock.Anything).
				Return(&donordata.Provided{}, nil)
			donorStore.EXPECT().
				Put(mock.Anything, mock.Anything).
				Return(expectedError)
			return donorStore
		},
	}

	for name, donorStore := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/lpa/id/path", nil)

			mux := http.NewServeMux()

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Sub: "random"}, nil)

			errorHandler := newMockErrorHandler(t)
			errorHandler.EXPECT().
				Execute(w, r, expectedError)

			handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore())
			handle("/path", page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *donordata.Provided) error {
				return expectedError
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

}

func TestMakeLpaHandleWhenCannotGoToURL(t *testing.T) {
	path := donor.PathWhenCanTheLpaBeUsed
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
		Return(&donordata.Provided{LpaID: "123", Donor: donordata.Donor{Email: "a@example.com"}}, nil)

	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle(path, page.None, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *donordata.Provided) error {
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("123"), resp.Header.Get("Location"))
}

func TestMakeLpaHandleSessionExistingSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "ignored"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/lpa/123/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&donordata.Provided{Donor: donordata.Donor{Email: "a@example.com"}}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle("/path", page.RequireSession|page.CanGoBack, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request, _ *donordata.Provided) error {
		assert.Equal(t, appcontext.Data{
			Page:      "/lpa/123/path",
			SessionID: "cmFuZG9t",
			CanGoBack: true,
			LpaID:     "123",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := appcontext.SessionFromContext(hr.Context())

		assert.Equal(t, &appcontext.Session{LpaID: "123", SessionID: "cmFuZG9t"}, sessionData)
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
		Return(&donordata.Provided{Donor: donordata.Donor{Email: "a@example.com"}}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore)
	handle("/path", page.RequireSession, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request, _ *donordata.Provided) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
