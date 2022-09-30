package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := Lpa{}

	dataStore := &mockDataStore{data: lpa}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere", Lpa: lpa}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Func, "/somewhere", dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestGuidanceWhenNilDataStore(t *testing.T) {
	w := httptest.NewRecorder()

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere"}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Func, "/somewhere", nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGuidanceWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := Lpa{}

	dataStore := &mockDataStore{data: lpa}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(nil, "/somewhere", dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere"}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Func, "/somewhere", dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}
