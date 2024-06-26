package app

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App(true, &slog.Logger{}, &localize.Localizer{}, localize.En, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &notify.Client{}, &place.Client{}, &onelogin.Client{}, nil, nil, nil, &search.Client{})

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page: "/path",
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
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

func TestMakeHandleRequireSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore)
	handle("/path", RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, errorHandler.Execute, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestWithAppData(t *testing.T) {
	testcases := map[string]struct {
		url                 string
		cookieName          string
		cookieConsentSet    bool
		showTranslationKeys bool
		contentType         string
	}{
		"with cookie consent": {
			url:              "/path?a=b",
			cookieName:       "cookies-consent",
			cookieConsentSet: true,
		},
		"without cookie consent": {
			url:        "/path?a=b",
			cookieName: "not-cookies-consent",
		},
		"with translation keys": {
			url:                 "/path?a=b&showTranslationKeys=1",
			showTranslationKeys: true,
		},
		"without translation keys": {
			url: "/path?a=b",
		},
		"with translation keys and multipart form": {
			url:                 "/path?a=b&showTranslationKeys=1",
			showTranslationKeys: false,
			contentType:         "multipart/form-data",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			r.AddCookie(&http.Cookie{Name: tc.cookieName, Value: "1"})
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}

			bundle, _ := localize.NewBundle("testdata/en.json")
			localizer := bundle.For(localize.En)
			localizer.SetShowTranslationKeys(tc.showTranslationKeys)

			query := url.Values{"a": {"b"}}
			if strings.Contains(tc.url, "showTranslationKeys") {
				query.Add("showTranslationKeys", "1")
			}

			handler := http.HandlerFunc(func(hw http.ResponseWriter, hr *http.Request) {
				assert.Equal(t, page.AppData{
					Path:             "/path",
					Query:            query,
					Localizer:        localizer,
					Lang:             localize.En,
					CookieConsentSet: tc.cookieConsentSet,
					CanToggleWelsh:   true,
				}, page.AppDataFromContext(hr.Context()))
				assert.Equal(t, w, hw)
			})

			withAppData(handler, localizer, localize.En)(w, r)
		})
	}
}
