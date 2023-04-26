package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardLpaData := []DashboardLpaDatum{
		{Lpa: &page.Lpa{ID: "123"}, CertificateProvider: &actor.CertificateProvider{}},
		{Lpa: &page.Lpa{ID: "456"}, CertificateProvider: &actor.CertificateProvider{}},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return([]*page.Lpa{{ID: "123"}, {ID: "456"}}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, nil)

	ctx = page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "456"})
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{App: testAppData, Lpas: dashboardLpaData}).
		Return(nil)

	err := Dashboard(template.Execute, lpaStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenCertificateProviderDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardLpaData := []DashboardLpaDatum{
		{Lpa: &page.Lpa{ID: "123"}, CertificateProvider: &actor.CertificateProvider{}},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return([]*page.Lpa{{ID: "123"}}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{App: testAppData, Lpas: dashboardLpaData}).
		Return(nil)

	err := Dashboard(template.Execute, lpaStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return([]*page.Lpa{}, expectedError)

	err := Dashboard(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return([]*page.Lpa{{ID: "123"}}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, expectedError)

	err := Dashboard(nil, lpaStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardLpaData := []DashboardLpaDatum{
		{Lpa: &page.Lpa{ID: "123"}, CertificateProvider: &actor.CertificateProvider{}},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("GetAll", r.Context()).
		Return([]*page.Lpa{{ID: "123"}}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})
	certificateProviderStore.
		On("Get", ctx).
		Return(&actor.CertificateProvider{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{App: testAppData, Lpas: dashboardLpaData}).
		Return(expectedError)

	err := Dashboard(template.Execute, lpaStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Create", r.Context()).
		Return(&page.Lpa{ID: "123"}, nil)

	err := Dashboard(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.YourDetails, resp.Header.Get("Location"))
}

func TestPostDashboardWhenLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Create", r.Context()).
		Return(&page.Lpa{ID: "123"}, expectedError)

	err := Dashboard(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
