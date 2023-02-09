package page

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const formUrlEncoded = "application/x-www-form-urlencoded"

var (
	expectedError = errors.New("err")
	appData       = AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      En,
		Paths:     Paths,
	}
)

type mockLpaStore struct {
	mock.Mock
}

func (m *mockLpaStore) Create(ctx context.Context) (*Lpa, error) {
	args := m.Called(ctx)

	return args.Get(0).(*Lpa), args.Error(1)
}

func (m *mockLpaStore) GetAll(ctx context.Context) ([]*Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Lpa), args.Error(1)
}

func (m *mockLpaStore) Get(ctx context.Context) (*Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).(*Lpa), args.Error(1)
}

func (m *mockLpaStore) Put(ctx context.Context, v *Lpa) error {
	return m.Called(ctx, v).Error(0)
}

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

func TestApp(t *testing.T) {
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &identity.YotiClient{}, "yoti-scenario-id", &notify.Client{}, &place.Client{}, RumConfig{}, "?%3fNEI0t9MN", appData.Paths, &onelogin.Client{})

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestCacheControlHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	CacheControlHeaders(http.NotFoundHandler()).ServeHTTP(w, r)

	resp := w.Result()
	assert.Equal(t, "max-age=2592000", resp.Header.Get("Cache-Control"))
}

func TestLangRedirect(t *testing.T) {
	testCases := map[Lang]string{
		En: "/dashboard",
		Cy: "/cy/dashboard",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, nil, "/dashboard")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestLangRedirectWhenLpaRoute(t *testing.T) {
	testCases := map[Lang]string{
		En: "/lpa/lpa-id/somewhere",
		Cy: "/cy/lpa/lpa-id/somewhere",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, nil, "/somewhere")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestLangRedirectWhenCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected string
	}{
		"nil": {
			lpa:      nil,
			expected: Paths.HowToConfirmYourIdentityAndSign,
		},
		"allowed": {
			lpa:      &Lpa{Tasks: Tasks{PayForLpa: TaskCompleted}},
			expected: Paths.HowToConfirmYourIdentityAndSign,
		},
		"not allowed": {
			lpa:      &Lpa{},
			expected: Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: En, LpaID: "lpa-id"}.Redirect(w, r, tc.lpa, Paths.HowToConfirmYourIdentityAndSign)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.expected, resp.Header.Get("Location"))
		})
	}
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, None)
	handle("/path", RequireSession|CanGoBack, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             En,
			SessionID:        "cmFuZG9t",
			CookieConsentSet: false,
			CanGoBack:        true,
			RumConfig:        RumConfig{ApplicationID: "xyz"},
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            AppPaths{},
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(contextWithSessionData(r.Context(), &sessionData{SessionID: "cmFuZG9t"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleRequireCertificateProviderSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Values: map[any]any{
				"certificate-provider": &CertificateProviderSession{
					Sub:       "random",
					SessionID: "session-id",
					LpaID:     "lpa-id",
				},
			},
		}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, None)
	handle("/path", RequireSession|RequireCertificateProvider, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             En,
			SessionID:        "session-id",
			LpaID:            "lpa-id",
			CookieConsentSet: false,
			CanGoBack:        false,
			RumConfig:        RumConfig{ApplicationID: "xyz"},
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            AppPaths{},
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(contextWithSessionData(r.Context(), &sessionData{SessionID: "session-id", LpaID: "lpa-id"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := contextWithSessionData(context.Background(), &sessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)
	localizer := localize.Localizer{}

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, None)
	handle("/path", RequireSession|CanGoBack, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             En,
			SessionID:        "cmFuZG9t",
			CookieConsentSet: false,
			CanGoBack:        true,
			RumConfig:        RumConfig{ApplicationID: "xyz"},
			StaticHash:       "?%3fNEI0t9MN",
			Paths:            AppPaths{},
			LpaID:            "123",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(contextWithSessionData(r.Context(), &sessionData{LpaID: "123", SessionID: "cmFuZG9t"})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleShowTranslationKeys(t *testing.T) {
	testCases := map[string]struct {
		showTranslationKeys string
		expected            bool
	}{
		"requested": {
			showTranslationKeys: "1",
			expected:            true,
		},
		"not requested": {
			showTranslationKeys: "maybe",
			expected:            false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/path?showTranslationKeys="+tc.showTranslationKeys, nil)
			localizer := localize.Localizer{}

			sessionsStore := &mockSessionsStore{}
			sessionsStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &DonorSession{Sub: "random"}}}, nil)

			mux := http.NewServeMux()
			handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, None)
			handle("/path", RequireSession|CanGoBack, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
				expectedLocalizer := localize.Localizer{}
				expectedLocalizer.ShowTranslationKeys = tc.expected

				assert.Equal(t, AppData{
					Page:             "/path",
					Query:            "?showTranslationKeys=" + tc.showTranslationKeys,
					Localizer:        expectedLocalizer,
					Lang:             En,
					SessionID:        "cmFuZG9t",
					CookieConsentSet: false,
					CanGoBack:        true,
					RumConfig:        RumConfig{ApplicationID: "xyz"},
					StaticHash:       "?%3fNEI0t9MN",
					Paths:            AppPaths{},
				}, appData)
				assert.Equal(t, w, hw)
				assert.Equal(t, r.WithContext(contextWithSessionData(r.Context(), &sessionData{SessionID: "cmFuZG9t"})), hr)
				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, sessionsStore)
		})
	}
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("Error rendering page for path '%s': %s", "/path", expectedError.Error()))

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{"donor": &DonorSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{}, None)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleSessionError(t *testing.T) {
	for name, opt := range map[string]handleOpt{
		"donor session":                RequireSession,
		"certificate provider session": RequireSession | RequireCertificateProvider,
	} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/path", nil)
			localizer := localize.Localizer{}

			logger := &mockLogger{}
			logger.
				On("Print", expectedError)

			sessionsStore := &mockSessionsStore{}
			sessionsStore.
				On("Get", r, "session").
				Return(&sessions.Session{}, expectedError)

			mux := http.NewServeMux()
			handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{Start: "/this"}, None)
			handle("/path", opt, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/this", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, sessionsStore, logger)
		})
	}
}

func TestMakeHandleSessionMissing(t *testing.T) {
	for name, tc := range map[string]struct {
		opt handleOpt
		err error
	}{
		"donor session": {
			opt: RequireSession,
			err: MissingSessionError("donor"),
		},
		"certificate provider session": {
			opt: RequireSession | RequireCertificateProvider,
			err: MissingSessionError("certificate-provider"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/path", nil)
			localizer := localize.Localizer{}

			logger := &mockLogger{}
			logger.
				On("Print", tc.err)

			sessionsStore := &mockSessionsStore{}
			sessionsStore.
				On("Get", r, "session").
				Return(&sessions.Session{Values: map[interface{}]interface{}{}}, nil)

			mux := http.NewServeMux()
			handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{Start: "/this"}, None)
			handle("/path", tc.opt, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/this", resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, sessionsStore, logger)
		})
	}
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{}, None)
	handle("/path", None, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Localizer:        localizer,
			Lang:             En,
			CookieConsentSet: false,
			StaticHash:       "?%3fNEI0t9MN",
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r, hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestTestingStart(t *testing.T) {
	t.Run("payment not complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{ID: "123"}).
			Return(nil)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("payment complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&paymentComplete=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:    "123",
				Tasks: Tasks{PayForLpa: TaskCompleted},
			}).
			Return(nil)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)
		sessionsStore.
			On("Get", r, "pay").
			Return(&sessions.Session{}, nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with payment", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withPayment=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:    "123",
				Tasks: Tasks{PayForLpa: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with attorney", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorney=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       "John@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				Tasks: Tasks{
					ChooseAttorneys: TaskCompleted,
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with incomplete attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompleteAttorneys=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		attorneys := actor.Attorneys{
			{
				ID:          "with-address",
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       "John@example.org",
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
			{
				ID:          "without-address",
				FirstNames:  "Joan",
				LastName:    "Smith",
				Email:       "Joan@example.org",
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{},
			},
		}

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                                   "123",
				Type:                                 LpaTypePropertyFinance,
				WhenCanTheLpaBeUsed:                  UsedWhenRegistered,
				Attorneys:                            attorneys,
				ReplacementAttorneys:                 attorneys,
				HowAttorneysMakeDecisions:            JointlyAndSeverally,
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				Tasks: Tasks{
					ChooseAttorneys:            TaskInProgress,
					ChooseReplacementAttorneys: TaskInProgress,
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorneys=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		attorneys := actor.Attorneys{
			{
				ID:          "JohnSmith",
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       "John@example.org",
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
			{
				ID:          "JoanSmith",
				FirstNames:  "Joan",
				LastName:    "Smith",
				Email:       "Joan@example.org",
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
		}

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                        "123",
				Attorneys:                 attorneys,
				HowAttorneysMakeDecisions: JointlyAndSeverally,
				Tasks: Tasks{
					ChooseAttorneys: TaskCompleted,
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("how attorneys act", func(t *testing.T) {
		testCases := []struct {
			DecisionsType    string
			DecisionsDetails string
		}{
			{DecisionsType: "jointly", DecisionsDetails: ""},
			{DecisionsType: "jointly-and-severally", DecisionsDetails: ""},
			{DecisionsType: "mixed", DecisionsDetails: "some details"},
		}

		for _, tc := range testCases {
			t.Run(tc.DecisionsType, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/?redirect=/somewhere&howAttorneysAct=%s", tc.DecisionsType), nil)
				ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

				sessionsStore := &mockSessionsStore{}
				sessionsStore.
					On("Save", r, w, mock.Anything).
					Return(nil)

				lpaStore := &mockLpaStore{}
				lpaStore.
					On("Create", ctx).
					Return(&Lpa{ID: "123"}, nil)
				lpaStore.
					On("Put", ctx, &Lpa{
						ID:                               "123",
						HowAttorneysMakeDecisions:        tc.DecisionsType,
						HowAttorneysMakeDecisionsDetails: tc.DecisionsDetails,
					}).
					Return(nil)

				testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
				mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
			})
		}
	})

	t.Run("with Certificate Provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withCP=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Barbara",
					LastName:                "Smith",
					Email:                   "Barbara@example.org",
					Mobile:                  "07535111111",
					DateOfBirth:             date.New("1997", "1", "2"),
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
				},
				Tasks: Tasks{CertificateProvider: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with donor details", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withDonorDetails=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				You: actor.Person{
					FirstNames: "Jose",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       "simulate-delivered@notifications.service.gov.uk",
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				Tasks:  Tasks{YourDetails: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with replacement attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withReplacementAttorneys=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                                   "123",
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				Tasks:                                Tasks{ChooseReplacementAttorneys: TaskCompleted},
				ReplacementAttorneys: actor.Attorneys{
					{
						FirstNames: "Jane",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       "Jane@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JaneSmith",
					},
					{
						FirstNames: "Jorge",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       "Jorge@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("when can be used completed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&whenCanBeUsedComplete=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                  "123",
				WhenCanTheLpaBeUsed: UsedWhenRegistered,
				Tasks:               Tasks{WhenCanTheLpaBeUsed: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with restrictions", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withRestrictions=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Restrictions: "Some restrictions on how Attorneys act",
				Tasks:        Tasks{Restrictions: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withPeopleToNotify=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				Tasks:                   Tasks{PeopleToNotify: TaskCompleted},
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      "Joanna@example.org",
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JonathanSmith",
						FirstNames: "Jonathan",
						LastName:   "Smith",
						Email:      "Jonathan@example.org",
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with incomplete people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompletePeopleToNotify=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      "Joanna@example.org",
						Address:    place.Address{},
					},
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("lpa checked", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&lpaChecked=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Checked:      true,
				HappyToShare: true,
				Tasks:        Tasks{CheckYourLpa: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("id confirmed and signed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&idConfirmedAndSigned=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				OneLoginUserData: identity.UserData{
					OK:          true,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FullName:    "Jose Smith",
				},
				WantToApplyForLpa:      true,
				WantToSignLpa:          true,
				CPWitnessCodeValidated: true,
				Submitted:              time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Tasks:                  Tasks{ConfirmYourIdentityAndSign: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("complete LPA", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&completeLpa=1", nil)
		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				OneLoginUserData: identity.UserData{
					OK:          true,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FullName:    "Jose Smith",
				},
				WantToApplyForLpa:       true,
				WantToSignLpa:           true,
				CPWitnessCodeValidated:  true,
				Submitted:               time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Checked:                 true,
				HappyToShare:            true,
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      "Joanna@example.org",
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JonathanSmith",
						FirstNames: "Jonathan",
						LastName:   "Smith",
						Email:      "Jonathan@example.org",
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				Restrictions:                         "Some restrictions on how Attorneys act",
				WhenCanTheLpaBeUsed:                  UsedWhenRegistered,
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				ReplacementAttorneys: actor.Attorneys{
					{
						FirstNames: "Jane",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       "Jane@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JaneSmith",
					},
					{
						FirstNames: "Jorge",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       "Jorge@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
				You: actor.Person{
					FirstNames: "Jose",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       "simulate-delivered@notifications.service.gov.uk",
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Barbara",
					LastName:                "Smith",
					Email:                   "Barbara@example.org",
					Mobile:                  "07535111111",
					DateOfBirth:             date.New("1997", "1", "2"),
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
				},
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       "John@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:          "JoanSmith",
						FirstNames:  "Joan",
						LastName:    "Smith",
						Email:       "Joan@example.org",
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				HowAttorneysMakeDecisions: JointlyAndSeverally,
				Tasks: Tasks{
					ConfirmYourIdentityAndSign: TaskCompleted,
					CheckYourLpa:               TaskCompleted,
					PeopleToNotify:             TaskCompleted,
					Restrictions:               TaskCompleted,
					WhenCanTheLpaBeUsed:        TaskCompleted,
					ChooseReplacementAttorneys: TaskCompleted,
					YourDetails:                TaskCompleted,
					CertificateProvider:        TaskCompleted,
					PayForLpa:                  TaskCompleted,
					ChooseAttorneys:            TaskCompleted,
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore, mockRandom).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})
}

func TestLangAbbreviation(t *testing.T) {
	type test struct {
		language string
		lang     Lang
		want     string
	}

	testCases := []test{
		{language: "English", lang: En, want: "en"},
		{language: "Welsh", lang: Cy, want: "cy"},
		{language: "Defaults to English with unsupported lang", lang: Lang(3), want: "en"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			a := tc.lang.String()
			assert.Equal(t, tc.want, a)
		})
	}
}

func TestLangBuildUrl(t *testing.T) {
	type test struct {
		language string
		lang     Lang
		url      string
		want     string
	}

	testCases := []test{
		{language: "English", lang: En, url: "/example.org", want: "/lpa/123/example.org"},
		{language: "Welsh", lang: Cy, url: "/example.org", want: "/cy/lpa/123/example.org"},
		{language: "Other", lang: Lang(3), url: "/example.org", want: "/lpa/123/example.org"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			builtUrl := AppData{Lang: tc.lang, LpaID: "123"}.BuildUrl(tc.url)
			assert.Equal(t, tc.want, builtUrl)
		})
	}
}

func TestRouteToLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/somewhere%2Fwhat", nil)

	routeToLpa(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/somewhere/what", r.URL.Path)
		assert.Equal(t, "/somewhere%2Fwhat", r.URL.RawPath)

		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(w, r)

	res := w.Result()

	assert.Equal(t, http.StatusTeapot, res.StatusCode)
}

func TestRouteToLpaWithoutID(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/", nil)

	routeToLpa(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(w, r)

	res := w.Result()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestQueryString(t *testing.T) {
	testCases := map[string]struct {
		url           string
		expectedQuery string
	}{
		"with query": {
			url:           "http://example.org/?a=query&b=string",
			expectedQuery: "?a=query&b=string",
		},
		"with empty query": {
			url:           "http://example.org/?",
			expectedQuery: "",
		},
		"without query": {
			url:           "http://example.org/",
			expectedQuery: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			assert.Equal(t, tc.expectedQuery, queryString(r))
		})
	}
}

func TestIsLpaPath(t *testing.T) {
	testCases := map[string]struct {
		url               string
		expectedIsLpaPage bool
	}{
		"dashboard": {
			url:               Paths.Dashboard + "?someQuery=5",
			expectedIsLpaPage: false,
		},
		"start": {
			url:               Paths.Start + "?someQuery=6",
			expectedIsLpaPage: false,
		},
		"any other page": {
			url:               "/other?someQuery=7",
			expectedIsLpaPage: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsLpaPage, IsLpaPath(tc.url))
		})
	}
}
