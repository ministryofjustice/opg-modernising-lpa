package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAboutPayment(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	template := &mockTemplate{}
	template.
		On("Func", w, &aboutPaymentData{App: appData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	err := AboutPayment(template.Func)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestAboutPaymentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	template := &mockTemplate{}
	template.
		On("Func", w, &aboutPaymentData{App: appData}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	err := AboutPayment(template.Func)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}
