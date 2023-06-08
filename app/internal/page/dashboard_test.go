package page

import (
	"context"
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

	donorLpas := []*Lpa{{ID: "123"}, {ID: "456"}}
	certificateProviderLpa := &Lpa{ID: "abc"}
	attorneyLpa := &Lpa{ID: "def"}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return(donorLpas, nil)
	donorStore.
		On("GetAny", ContextWithSessionData(context.Background(), &SessionData{LpaID: "abc"})).
		Return(certificateProviderLpa, nil)
	donorStore.
		On("GetAny", ContextWithSessionData(context.Background(), &SessionData{LpaID: "def"})).
		Return(attorneyLpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAll", r.Context()).
		Return([]*actor.CertificateProviderProvidedDetails{{
			LpaID: "abc",
		}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAll", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{{
			LpaID: "def",
		}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{
			App:                     AppData{},
			UseTabs:                 true,
			DonorLpas:               donorLpas,
			AttorneyLpas:            []*Lpa{attorneyLpa},
			CertificateProviderLpas: []*Lpa{certificateProviderLpa},
		}).
		Return(nil)

	err := Dashboard(template.Execute, donorStore, certificateProviderStore, attorneyStore)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardOnlyDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []*Lpa{{ID: "123"}, {ID: "456"}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return(donorLpas, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAll", r.Context()).
		Return([]*actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAll", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &dashboardData{
			App:                     AppData{},
			DonorLpas:               donorLpas,
			AttorneyLpas:            []*Lpa{},
			CertificateProviderLpas: []*Lpa{},
		}).
		Return(nil)

	err := Dashboard(template.Execute, donorStore, certificateProviderStore, attorneyStore)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return([]*Lpa{}, expectedError)

	err := Dashboard(nil, donorStore, nil, nil)(AppData{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []*Lpa{{ID: "123"}, {ID: "456"}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return(donorLpas, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAll", r.Context()).
		Return(nil, expectedError)

	err := Dashboard(nil, donorStore, certificateProviderStore, nil)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestGetDashboardWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []*Lpa{{ID: "123"}, {ID: "456"}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return(donorLpas, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAll", r.Context()).
		Return([]*actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAll", r.Context()).
		Return(nil, expectedError)

	err := Dashboard(nil, donorStore, certificateProviderStore, attorneyStore)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestGetDashboardWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []*Lpa{{ID: "123"}, {ID: "456"}}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAll", r.Context()).
		Return(donorLpas, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAll", r.Context()).
		Return([]*actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAll", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := Dashboard(template.Execute, donorStore, certificateProviderStore, attorneyStore)(AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Create", r.Context()).
		Return(&Lpa{ID: "lpa-id"}, nil)

	err := Dashboard(nil, donorStore, nil, nil)(AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.YourDetails, resp.Header.Get("Location"))
}

func TestPostDashboardWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Create", r.Context()).
		Return(&Lpa{ID: "123"}, expectedError)

	err := Dashboard(nil, donorStore, nil, nil)(AppData{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
