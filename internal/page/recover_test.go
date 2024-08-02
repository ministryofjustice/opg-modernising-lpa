package page

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
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
			template.EXPECT().
				Execute(w, &errorData{App: appcontext.Data{
					CookieConsentSet: true,
					Lang:             lang,
					Localizer:        (*localize.Localizer)(nil),
				}}).
				Return(nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				ErrorContext(r.Context(), "recover error",
					slog.Any("req", r),
					mock.MatchedBy(func(a slog.Attr) bool {
						return assert.ErrorContains(t, a.Value.Any().(error), "runtime error")
					}),
					mock.MatchedBy(func(a slog.Attr) bool {
						return assert.Contains(t, a.Value.String(), "github.com/ministryofjustice/opg-modernising-lpa/internal/page.TestRecover") &&
							assert.Contains(t, a.Value.String(), "internal/page/recover_test.go:")
					}))

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(lang).
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
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(r.Context(), "recover error", mock.Anything, mock.Anything, mock.Anything)
	logger.EXPECT().
		ErrorContext(r.Context(), "error rendering page", slog.Any("req", r), slog.Any("err", expectedError))

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
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
