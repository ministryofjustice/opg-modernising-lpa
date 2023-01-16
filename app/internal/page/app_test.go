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
		Lang:      En,
		Paths:     Paths,
	}
)

type mockLpaStore struct {
	mock.Mock
}

func (m *mockLpaStore) Get(ctx context.Context, sessionID string) (*Lpa, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*Lpa), args.Error(1)
}

func (m *mockLpaStore) Put(ctx context.Context, id string, v *Lpa) error {
	return m.Called(ctx, id, v).Error(0)
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
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &identity.YotiClient{}, "yoti-scenario-id", &notify.Client{}, &place.Client{}, RumConfig{}, "?%3fNEI0t9MN", appData.Paths, &onelogin.Client{}, false)

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
		En: "/somewhere",
		Cy: "/cy/somewhere",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			lang.Redirect(w, r, nil, "/somewhere")
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

			En.Redirect(w, r, tc.lpa, Paths.HowToConfirmYourIdentityAndSign)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
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
		Return(&sessions.Session{Values: map[interface{}]interface{}{"sub": "random"}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, false)
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
		assert.Equal(t, r, hr)
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
		isProduction        bool
		showTranslationKeys string
		expected            bool
	}{
		"enabled": {
			isProduction:        false,
			showTranslationKeys: "1",
			expected:            true,
		},
		"disabled production": {
			isProduction:        true,
			showTranslationKeys: "1",
			expected:            false,
		},
		"disabled not requested": {
			isProduction:        false,
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
				Return(&sessions.Session{Values: map[interface{}]interface{}{"sub": "random"}}, nil)

			mux := http.NewServeMux()
			handle := makeHandle(mux, nil, sessionsStore, localizer, En, RumConfig{ApplicationID: "xyz"}, "?%3fNEI0t9MN", AppPaths{}, tc.isProduction)
			handle("/path", RequireSession|CanGoBack, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
				expectedLocalizer := localize.Localizer{}
				expectedLocalizer.ShowTranslationKeys = tc.expected

				assert.Equal(t, AppData{
					Page:                "/path",
					Query:               "?showTranslationKeys=" + tc.showTranslationKeys,
					Localizer:           expectedLocalizer,
					Lang:                En,
					SessionID:           "cmFuZG9t",
					CookieConsentSet:    false,
					CanGoBack:           true,
					RumConfig:           RumConfig{ApplicationID: "xyz"},
					StaticHash:          "?%3fNEI0t9MN",
					Paths:               AppPaths{},
					DevFeaturesEnabled:  tc.isProduction,
					ShowTranslationKeys: tc.showTranslationKeys == "1",
				}, appData)
				assert.Equal(t, w, hw)
				assert.Equal(t, r, hr)
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
		Return(&sessions.Session{Values: map[interface{}]interface{}{"sub": "random"}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{}, false)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, sessionsStore)
}

func TestMakeHandleSessionError(t *testing.T) {
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
	handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{Start: "/this"}, false)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/this", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", "sub missing from session")

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[interface{}]interface{}{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, logger, sessionsStore, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{Start: "/this"}, false)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/this", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, localizer, En, RumConfig{}, "?%3fNEI0t9MN", AppPaths{}, false)
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

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		testingStart(sessionsStore, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("payment complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&paymentComplete=1", nil)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)
		sessionsStore.
			On("Get", r, "pay").
			Return(&sessions.Session{}, nil)

		testingStart(sessionsStore, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with payment", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withPayment=1", nil)
		r = r.WithContext(ctx)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Get", ctx, mock.Anything).
			Return(&Lpa{}, nil)
		lpaStore.
			On("Put", ctx, mock.Anything, &Lpa{
				Tasks: Tasks{PayForLpa: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with attorney", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorney=1", nil)
		r = r.WithContext(ctx)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Get", ctx, mock.Anything).
			Return(&Lpa{}, nil)
		lpaStore.
			On("Put", ctx, mock.Anything, &Lpa{
				Attorneys: []Attorney{
					{
						ID:          "with-address",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       "aa@example.org",
						DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
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
					ChooseAttorneys: TaskCompleted,
				},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with attorneys", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompleteAttorneys=1", nil)
		r = r.WithContext(ctx)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Get", ctx, mock.Anything).
			Return(&Lpa{}, nil)

		attorneys := []Attorney{
			{
				ID:          "with-address",
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       "aa@example.org",
				DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
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
				Email:       "bb@example.org",
				DateOfBirth: time.Date(1998, time.January, 2, 3, 4, 5, 6, time.UTC),
				Address:     place.Address{},
			},
		}

		lpaStore.
			On("Put", ctx, mock.Anything, &Lpa{
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

		testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
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
				ctx := context.Background()
				w := httptest.NewRecorder()
				r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/?redirect=/somewhere&howAttorneysAct=%s", tc.DecisionsType), nil)
				r = r.WithContext(ctx)

				sessionsStore := &mockSessionsStore{}
				sessionsStore.
					On("Get", r, "session").
					Return(&sessions.Session{}, nil)
				sessionsStore.
					On("Save", r, w, mock.Anything).
					Return(nil)

				lpaStore := &mockLpaStore{}
				lpaStore.
					On("Get", ctx, mock.Anything).
					Return(&Lpa{}, nil)

				lpaStore.
					On("Put", ctx, mock.Anything, &Lpa{HowAttorneysMakeDecisions: tc.DecisionsType, HowAttorneysMakeDecisionsDetails: tc.DecisionsDetails}).
					Return(nil)

				testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
				mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
			})
		}
	})

	t.Run("with Certificate Provider", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withCP=1", nil)
		r = r.WithContext(ctx)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Get", r, "session").
			Return(&sessions.Session{}, nil)
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &mockLpaStore{}
		lpaStore.
			On("Get", ctx, mock.Anything).
			Return(&Lpa{}, nil)

		lpaStore.
			On("Put", ctx, mock.Anything, &Lpa{
				CertificateProvider: CertificateProvider{
					FirstNames:              "Barbara",
					LastName:                "Smith",
					Email:                   "b@example.org",
					Mobile:                  "07535111111",
					DateOfBirth:             time.Date(1997, time.January, 2, 3, 4, 5, 6, time.UTC),
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
				},
				Tasks: Tasks{CertificateProvider: TaskCompleted},
			}).
			Return(nil)

		testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
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
		{language: "English", lang: En, url: "/example.org", want: "/example.org"},
		{language: "Welsh", lang: Cy, url: "/example.org", want: "/cy/example.org"},
		{language: "Other", lang: Lang(3), url: "/example.org", want: "/example.org"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			builtUrl := tc.lang.BuildUrl(tc.url)
			assert.Equal(t, tc.want, builtUrl)
		})
	}
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
