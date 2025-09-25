package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
	testNow       = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn     = func() time.Time { return testNow }
)

func (m *mockDynamoClient) ExpectOne(ctx, pk, sk, data interface{}, err error) {
	m.
		On("One", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectAllBySK(ctx, sk, data interface{}, err error) {
	m.
		On("AllBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByKeys(ctx context.Context, keys []dynamo.Keys, data []map[string]types.AttributeValue, err error) {
	m.EXPECT().
		AllByKeys(ctx, keys).
		Return(data, err)
}

func (m *mockDynamoClient) ExpectOneBySK(ctx, sk, data interface{}, err error) {
	m.
		On("OneBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByPartialSK(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("AllByPartialSK", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectOneByPK(ctx, pk, data interface{}, err error) {
	m.
		On("OneByPK", ctx, pk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func TestApp(t *testing.T) {
	app := App(true, &slog.Logger{}, &localize.Bundle{}, localize.En, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, template.Templates{}, nil, nil, "http://public.url", &pay.Client{}, &notify.Client{}, &place.Client{}, &onelogin.Client{}, nil, nil, nil, &search.Client{}, "http://use.url", "http://donor.url", "http://certificate.url", "http://attorney.url", true)

	assert.Implements(t, (*http.Handler)(nil), app)
}

func TestMakeHandle(t *testing.T) {
	testcases := map[string]struct {
		opts                handleOpt
		expectedData        appcontext.Data
		expectedSessionData *appcontext.Session
		setupSessionStore   func(*http.Request) *mockSessionStore
	}{
		"no opts": {
			opts: None,
			expectedData: appcontext.Data{
				Page: "/path",
			},
			setupSessionStore: func(*http.Request) *mockSessionStore { return newMockSessionStore(t) },
		},
		"RequireSession": {
			opts: RequireSession,
			expectedData: appcontext.Data{
				Page:      "/path",
				SessionID: "cmFuZG9t",
			},
			expectedSessionData: &appcontext.Session{SessionID: "cmFuZG9t"},
			setupSessionStore: func(r *http.Request) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					Login(r).
					Return(&sesh.LoginSession{Sub: "random"}, nil)
				return sessionStore
			},
		},
		"HideNav": {
			opts: HideNav,
			expectedData: appcontext.Data{
				Page:         "/path",
				HideLoginNav: true,
			},
			setupSessionStore: func(*http.Request) *mockSessionStore { return newMockSessionStore(t) },
		},
		"RequireSession|HideNave": {
			opts: RequireSession | HideNav,
			expectedData: appcontext.Data{
				Page:         "/path",
				SessionID:    "cmFuZG9t",
				HideLoginNav: true,
			},
			expectedSessionData: &appcontext.Session{SessionID: "cmFuZG9t"},
			setupSessionStore: func(r *http.Request) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					Login(r).
					Return(&sesh.LoginSession{Sub: "random"}, nil)
				return sessionStore
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

			mux := http.NewServeMux()
			handle := makeHandle(mux, nil, tc.setupSessionStore(r), "http://example.com/donor")
			handle("/path", tc.opts, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
				assert.Equal(t, tc.expectedData, appData)
				assert.Equal(t, w, hw)

				sessionData, _ := appcontext.SessionFromContext(hr.Context())
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

func TestMakeHandleRequireSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, sessionStore, "http://example.com/donor")
	handle("/path", RequireSession, func(_ appcontext.Data, _ http.ResponseWriter, _ *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://example.com/donor", resp.Header.Get("Location"))
}

func TestMakeHandleWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, errorHandler.Execute, nil, "http://example.com/donor")
	handle("/path", None, func(appData appcontext.Data, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestWithAppData(t *testing.T) {
	testcases := map[string]struct {
		url              string
		cookieName       string
		cookieValue      string
		cookieConsentSet bool
	}{
		"with cookie consent": {
			url:              "/path?a=b",
			cookieName:       "cookies-consent",
			cookieValue:      "1",
			cookieConsentSet: true,
		},
		"without cookie consent": {
			url:        "/path?a=b",
			cookieName: "not-cookies-consent",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			r.AddCookie(&http.Cookie{Name: tc.cookieName, Value: tc.cookieValue})

			bundle, _ := localize.NewBundle("testdata/en.json")
			localizer := bundle.For(localize.En)

			query := url.Values{"a": {"b"}}

			handler := http.HandlerFunc(func(hw http.ResponseWriter, hr *http.Request) {
				assert.Equal(t, appcontext.Data{
					Path:             "/path",
					Query:            query,
					Localizer:        localizer,
					Lang:             localize.En,
					CookieConsentSet: tc.cookieConsentSet,
				}, appcontext.DataFromContext(hr.Context()))
				assert.Equal(t, w, hw)
			})

			withAppData(handler, localizer, localize.En, false)(w, r)
		})
	}
}

func TestWithAppDataWithDevMode(t *testing.T) {
	testcases := map[string]struct {
		url                 string
		cookieName          string
		cookieValue         string
		cookieConsentSet    bool
		showTranslationKeys bool
		contentType         string
		devMode             bool
	}{
		"with translation keys": {
			url:                 "/path?a=b&showTranslationKeys=1",
			showTranslationKeys: true,
			devMode:             true,
		},
		"with translation keys, dev mode off": {
			url: "/path?a=b&showTranslationKeys=1",
		},
		"without translation keys": {
			url:     "/path?a=b",
			devMode: true,
		},
		"without translation keys, dev mode off": {
			url: "/path?a=b",
		},
		"with translation keys cookie": {
			url:                 "/path?a=b",
			cookieName:          "show-keys",
			cookieValue:         "1",
			showTranslationKeys: true,
			devMode:             true,
		},
		"with translation keys cookie, dev mode off": {
			url:         "/path?a=b",
			cookieName:  "show-keys",
			cookieValue: "1",
		},
		"disable translation keys cookie": {
			url:                 "/path?a=b&showTranslationKeys=0",
			cookieName:          "show-keys",
			cookieValue:         "0",
			showTranslationKeys: false,
			devMode:             true,
		},
		"with translation keys and multipart form": {
			url:                 "/path?a=b&showTranslationKeys=1",
			showTranslationKeys: false,
			contentType:         "multipart/form-data",
			devMode:             true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			r.AddCookie(&http.Cookie{Name: tc.cookieName, Value: tc.cookieValue})
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}

			bundle, _ := localize.NewBundle("testdata/en.json")
			localizer := bundle.For(localize.En)
			localizer.SetShowTranslationKeys(tc.showTranslationKeys)

			query := url.Values{"a": {"b"}}
			if strings.Contains(tc.url, "showTranslationKeys") {
				value := strings.SplitAfter(tc.url, "showTranslationKeys=")
				query.Add("showTranslationKeys", value[1])
			}

			handler := http.HandlerFunc(func(hw http.ResponseWriter, hr *http.Request) {
				assert.Equal(t, appcontext.Data{
					Path:             "/path",
					Query:            query,
					Localizer:        localizer,
					Lang:             localize.En,
					CookieConsentSet: tc.cookieConsentSet,
				}, appcontext.DataFromContext(hr.Context()))
				assert.Equal(t, w, hw)
			})

			withAppData(handler, localizer, localize.En, true)(w, r)
		})
	}
}
