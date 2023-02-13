package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{ID: "123"}, {ID: "456"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpas: lpas}).
		Return(nil)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetDashboardWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, expectedError)

	err := Dashboard(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &dashboardData{App: appData, Lpas: lpas}).
		Return(expectedError)

	err := Dashboard(template.Func, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Create", r.Context()).
		Return(&page.Lpa{ID: "123"}, nil)

	err := Dashboard(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YourDetails, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}
