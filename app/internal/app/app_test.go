package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := App(&logging.Logger{}, &localize.Localizer{}, localize.En, template.Templates{}, nil, &dynamo.Client{}, "http://public.url", &pay.Client{}, &identity.YotiClient{}, "yoti-donor-scenario-id", "yoti-certificate-provider-scenario-id", &notify.Client{}, &place.Client{}, page.RumConfig{}, "?%3fNEI0t9MN", page.Paths, &onelogin.Client{})

	assert.Implements(t, (*http.Handler)(nil), app)
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

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil)
	handle("/path", func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
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

func TestMakeHandleWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, errorHandler.Execute)
	handle("/path", func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}
