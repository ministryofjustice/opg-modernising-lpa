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

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const formUrlEncoded = "application/x-www-form-urlencoded"

var (
	expectedError = errors.New("err")
	appData       = AppData{SessionID: "session-id", Lang: En}
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
	app := App(&mockLogger{}, localize.Localizer{}, En, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &identity.YotiClient{}, "yoti-scenario-id", &notify.Client{}, &place.Client{})

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestLangRedirect(t *testing.T) {
	testCases := map[Lang]string{
		En: "/somewhere",
		Cy: "/cy/somewhere",
	}

	for lang, url := range testCases {
		t.Run("En", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			lang.Redirect(w, r, "/somewhere", http.StatusFound)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
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
	handle := makeHandle(mux, nil, sessionsStore, localizer, En)
	handle("/path", RequireSession|CanGoBack, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Query:            "?a=b",
			Localizer:        localizer,
			Lang:             En,
			SessionID:        "cmFuZG9t",
			CookieConsentSet: false,
			CanGoBack:        true,
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
	handle := makeHandle(mux, logger, sessionsStore, localizer, En)
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
	handle := makeHandle(mux, logger, sessionsStore, localizer, En)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
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
	handle := makeHandle(mux, logger, sessionsStore, localizer, En)
	handle("/path", RequireSession, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, startPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, sessionsStore, logger)
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	localizer := localize.Localizer{}

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil, localizer, En)
	handle("/path", None, func(appData AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, AppData{
			Page:             "/path",
			Localizer:        localizer,
			Lang:             En,
			CookieConsentSet: false,
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

	t.Run("with attorneys", func(t *testing.T) {
		ctx := context.Background()
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorneys=1", nil)
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

		updatedLpa := &Lpa{}
		attorneys := []Attorney{
			{
				ID:          "completed-address",
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
				ID:          "empty-address",
				FirstNames:  "Joan",
				LastName:    "Smith",
				Email:       "bb@example.org",
				DateOfBirth: time.Date(1998, time.January, 2, 3, 4, 5, 6, time.UTC),
				Address:     place.Address{},
			},
		}

		updatedLpa.Attorneys = attorneys
		updatedLpa.ReplacementAttorneys = attorneys

		lpaStore.
			On("Put", ctx, mock.Anything, updatedLpa).
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
					On("Put", ctx, mock.Anything, &Lpa{DecisionsType: tc.DecisionsType, DecisionsDetails: tc.DecisionsDetails}).
					Return(nil)

				testingStart(sessionsStore, lpaStore).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/somewhere", resp.Header.Get("Location"))
				mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
			})
		}
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
