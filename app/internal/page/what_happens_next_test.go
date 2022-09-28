package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWhatHappensNext(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	template := &mockTemplate{}
	template.
		On("Func", w, &whatHappensNextData{App: appData, Continue: taskListPath}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhatHappensNext(template.Func)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestWhatHappensNextWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	appData := AppData{}

	template := &mockTemplate{}
	template.
		On("Func", w, &whatHappensNextData{App: appData, Continue: taskListPath}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhatHappensNext(template.Func)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}
