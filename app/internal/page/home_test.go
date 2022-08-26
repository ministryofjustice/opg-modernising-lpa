package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHome(t *testing.T) {
	w := httptest.NewRecorder()

	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, &homeData{Page: homePath, L: localizer, Lang: En, SignInURL: "/here"}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Home(nil, localizer, En, template.Func, "/here")(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestHomeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)
	template := &mockTemplate{}
	template.
		On("Func", w, &homeData{Page: homePath, L: localizer, Lang: En, SignInURL: "/here"}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Home(logger, localizer, En, template.Func, "/here")(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, logger)
}
