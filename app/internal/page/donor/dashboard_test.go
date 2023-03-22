package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{ID: "123"}, {ID: "456"}}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{App: testAppData, Lpas: lpas}).
		Return(nil)

	err := Dashboard(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{}}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, expectedError)

	err := Dashboard(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpas := []*page.Lpa{{}}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return(lpas, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{App: testAppData, Lpas: lpas}).
		Return(expectedError)

	err := Dashboard(template.Execute, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostDashboardCreate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Create", r.Context()).
		Return(&page.Lpa{ID: "123"}, nil)

	err := Dashboard(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YourDetails, resp.Header.Get("Location"))
}

func TestPostDashboardCreateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Create", r.Context()).
		Return(nil, expectedError)

	err := Dashboard(nil, lpaStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostDashboardReuse(t *testing.T) {
	form := url.Values{
		"action":   {"reuse"},
		"reuse-id": {"123"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Clone", r.Context(), "123").
		Return(&page.Lpa{ID: "123"}, nil)

	err := Dashboard(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YourDetails, resp.Header.Get("Location"))
}

func TestPostDashboardReuseErrors(t *testing.T) {
	form := url.Values{
		"action":   {"reuse"},
		"reuse-id": {"123"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Clone", r.Context(), "123").
		Return(nil, expectedError)

	err := Dashboard(nil, lpaStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}
