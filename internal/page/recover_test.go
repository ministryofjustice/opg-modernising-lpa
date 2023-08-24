package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecover(t *testing.T) {
	testcases := map[localize.Lang]string{
		localize.En: "/",
		localize.Cy: "/cy/",
	}

	for lang, url := range testcases {
		t.Run(lang.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, url, nil)

			badHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var uninitialised http.Handler
				uninitialised.ServeHTTP(w, r)
			})

			template := newMockTemplate(t)
			template.
				On("Execute", w, &errorData{App: AppData{
					CookieConsentSet: true,
					Paths:            Paths,
					Lang:             lang,
					Localizer:        (*localize.Localizer)(nil),
				}}).
				Return(nil)

			logger := newMockLogger(t)
			logger.
				On("Request", r, mock.MatchedBy(func(e recoverError) bool {
					return assert.Equal(t, "recover error", e.Error()) &&
						assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", e.Title()) &&
						assert.Contains(t, e.Data(), "github.com/ministryofjustice/opg-modernising-lpa/internal/page.TestRecover") &&
						assert.Contains(t, e.Data(), "internal/page/recover_test.go:")
				}))

			bundle := newMockBundle(t)
			bundle.
				On("For", lang).
				Return(nil)

			Recover(template.Execute, logger, bundle, badHandler).ServeHTTP(w, r)
		})
	}
}

func TestRecoverWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	badHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uninitialised http.Handler
		uninitialised.ServeHTTP(w, r)
	})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(ExpectedError)

	logger := newMockLogger(t)
	logger.
		On("Request", r, mock.Anything)
	logger.
		On("Print", "Error rendering page: err")

	bundle := newMockBundle(t)
	bundle.
		On("For", mock.Anything).
		Return(nil)

	Recover(template.Execute, logger, bundle, badHandler).ServeHTTP(w, r)
}

func TestRecoverWhenNoPanic(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	goodHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	Recover(nil, nil, nil, goodHandler).ServeHTTP(w, r)
	resp := w.Result()
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}
