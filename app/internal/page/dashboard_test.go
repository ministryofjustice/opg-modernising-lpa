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
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpa: lpa}).
		Return(nil)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestDashboardWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpa, expectedError)

	err := Dashboard(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpa: &Lpa{}}).
		Return(expectedError)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
