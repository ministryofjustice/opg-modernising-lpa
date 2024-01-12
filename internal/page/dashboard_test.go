package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []LpaAndActorTasks{
		{Donor: &actor.DonorProvidedDetails{LpaID: "123"}},
		{Donor: &actor.DonorProvidedDetails{LpaID: "456"}},
	}

	certificateProviderLpas := []LpaAndActorTasks{{Donor: &actor.DonorProvidedDetails{LpaID: "abc"}}}
	attorneyLpas := []LpaAndActorTasks{{Donor: &actor.DonorProvidedDetails{LpaID: "def"}}}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(donorLpas, attorneyLpas, certificateProviderLpas, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:                     AppData{},
			UseTabs:                 true,
			DonorLpas:               donorLpas,
			AttorneyLpas:            attorneyLpas,
			CertificateProviderLpas: certificateProviderLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, nil, dashboardStore)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardOnlyDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []LpaAndActorTasks{
		{Donor: &actor.DonorProvidedDetails{LpaID: "123"}},
		{Donor: &actor.DonorProvidedDetails{LpaID: "456"}},
	}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(donorLpas, nil, nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:       AppData{},
			DonorLpas: donorLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, nil, dashboardStore)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenDashboardStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(nil, nil, nil, expectedError)

	err := Dashboard(nil, nil, dashboardStore)(AppData{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(nil, nil, nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Dashboard(template.Execute, nil, dashboardStore)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Create(r.Context()).
		Return(&actor.DonorProvidedDetails{LpaID: "lpa-id"}, nil)

	err := Dashboard(nil, donorStore, nil)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.YourDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostDashboardWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Create(r.Context()).
		Return(&actor.DonorProvidedDetails{LpaID: "123"}, expectedError)

	err := Dashboard(nil, donorStore, nil)(AppData{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
