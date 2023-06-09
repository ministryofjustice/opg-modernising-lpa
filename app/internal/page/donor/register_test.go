package donor

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &log.Logger{}, template.Templates{}, nil, nil, &onelogin.Client{}, &place.Client{}, "http://public.url", &pay.Client{}, &identity.YotiClient{}, &notify.Client{}, nil, nil, nil, nil, &uid.Client{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			CanGoBack: true,
			SessionID: "cmFuZG9t",
			ActorType: actor.TypeDonor,
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

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
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

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, errorHandler.Execute)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, None, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(page.ContextWithAppData(r.Context(), page.AppData{Page: "/path", ActorType: actor.TypeDonor})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeLpaHandleWhenDetailsProvidedAndUIDExists(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
			UID:   "a-uid",
		}, nil)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, donorStore, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
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

func TestMakeLpaHandleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, nil, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start, resp.Header.Get("Location"))
}

func TestMakeLpaHandleWhenLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, errorHandler.Execute, donorStore, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeLpaHandleWhenDetailsProvidedAndUIDDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}, nil)

	donorStore.
		On("Put", mock.Anything, &page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
			UID:   "M-789Q-P4DF-4UX3",
		}).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", mock.Anything, &uid.CreateCaseRequestBody{
			Type: page.LpaTypePropertyFinance,
			Donor: uid.DonorDetails{
				Name:     "Jane Smith",
				Dob:      uid.ISODate{Time: date.New("2000", "1", "2").Time()},
				Postcode: "ABC123",
			},
		}).
		Return(uid.CreateCaseResponse{UID: "M-789Q-P4DF-4UX3"}, nil)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, donorStore, uidClient, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
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

func TestMakeLpaHandleWhenDetailsProvidedAndUIDDoesNotExistOnUidClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}, nil)

	donorStore.
		On("Put", mock.Anything, &page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", mock.Anything, &uid.CreateCaseRequestBody{
			Type: page.LpaTypePropertyFinance,
			Donor: uid.DonorDetails{
				Name:     "Jane Smith",
				Dob:      uid.ISODate{Time: date.New("2000", "1", "2").Time()},
				Postcode: "ABC123",
			},
		}).
		Return(uid.CreateCaseResponse{}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError).
		Return(nil)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, donorStore, uidClient, logger)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
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

func TestMakeLpaHandleWhenDetailsProvidedAndUIDDoesNotExistOnLpaStorePutError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}, nil)

	donorStore.
		On("Put", mock.Anything, &page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}).
		Return(expectedError)

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", mock.Anything, &uid.CreateCaseRequestBody{
			Type: page.LpaTypePropertyFinance,
			Donor: uid.DonorDetails{
				Name:     "Jane Smith",
				Dob:      uid.ISODate{Time: date.New("2000", "1", "2").Time()},
				Postcode: "ABC123",
			},
		}).
		Return(uid.CreateCaseResponse{}, nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, errorHandler.Execute, donorStore, uidClient, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeLpaHandleSessionExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, None, nil, donorStore, nil, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
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
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, None, errorHandler.Execute, donorStore, nil, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
