package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpa: lpa}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestDashboardWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Dashboard(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpa: &Lpa{}}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
