package page

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "123"}}}
	registeredDonorLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "456", RegisteredAt: time.Now()}}}
	certificateProviderLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "abc"}}}
	attorneyLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "def"}}}
	registeredAttorneyLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "xyz", RegisteredAt: time.Now()}}}
	voucherLpas := []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "def"}}}

	results := dashboarddata.Results{
		Donor:               append(donorLpas, registeredDonorLpas...),
		CertificateProvider: certificateProviderLpas,
		Attorney:            append(attorneyLpas, registeredAttorneyLpas...),
		Voucher:             voucherLpas,
	}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(results, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:                     appcontext.Data{},
			NeedsTabs:               true,
			DonorLpas:               donorLpas,
			RegisteredDonorLpas:     registeredDonorLpas,
			AttorneyLpas:            attorneyLpas,
			RegisteredAttorneyLpas:  registeredAttorneyLpas,
			CertificateProviderLpas: certificateProviderLpas,
			VoucherLpas:             voucherLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, dashboardStore, "", nil)(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardOnlyDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []dashboarddata.Actor{
		{Lpa: &lpadata.Lpa{LpaID: "123"}},
		{Lpa: &lpadata.Lpa{LpaID: "456"}},
	}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{Donor: donorLpas}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:       appcontext.Data{},
			DonorLpas: donorLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, dashboardStore, "", nil)(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardNoResults(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{HasLPAs: true, Email: "a@example.com"}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{Email: "a@example.com"}).
		Return(nil)

	err := Dashboard(nil, dashboardStore, "", sessionStore)(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, PathMakeOrAddAnLPA.String(), resp.Header.Get("Location"))
}

func TestGetDashboardNoResultsWhenSessionGetError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(mock.Anything).
		Return(dashboarddata.Results{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(nil, expectedError)

	err := Dashboard(nil, dashboardStore, "", sessionStore)(appcontext.Data{}, w, r)

	assert.ErrorIs(t, expectedError, err)
}

func TestGetDashboardNoResultsWhenSessionSetError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(mock.Anything).
		Return(dashboarddata.Results{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := Dashboard(nil, dashboardStore, "", sessionStore)(appcontext.Data{}, w, r)

	assert.ErrorIs(t, expectedError, err)

}

func TestGetDashboardWhenDashboardStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{}, expectedError)

	err := Dashboard(nil, dashboardStore, "", nil)(appcontext.Data{}, w, r)
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
		Return(dashboarddata.Results{Donor: []dashboarddata.Actor{{Lpa: &lpadata.Lpa{LpaID: "123"}}}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Dashboard(template.Execute, dashboardStore, "", nil)(appcontext.Data{}, w, r)
	assert.Equal(t, expectedError, err)
}
