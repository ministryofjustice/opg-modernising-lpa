package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []LpaAndActorTasks{
		{Lpa: &lpastore.Lpa{LpaID: "123"}},
		{Lpa: &lpastore.Lpa{LpaID: "456"}},
	}

	certificateProviderLpas := []LpaAndActorTasks{{Lpa: &lpastore.Lpa{LpaID: "abc"}}}
	attorneyLpas := []LpaAndActorTasks{{Lpa: &lpastore.Lpa{LpaID: "def"}}}

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
		{Lpa: &lpastore.Lpa{LpaID: "123"}},
		{Lpa: &lpastore.Lpa{LpaID: "456"}},
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
	testCases := map[string]struct {
		Form             url.Values
		ExpectedRedirect LpaPath
	}{
		"no donor LPAs": {
			ExpectedRedirect: Paths.YourDetails,
		},
		"with donor LPAs": {
			Form:             url.Values{"has-existing-donor-lpas": {"true"}},
			ExpectedRedirect: Paths.MakeANewLPA,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.Form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Create(r.Context()).
				Return(&actor.DonorProvidedDetails{LpaID: "lpa-id"}, nil)

			err := Dashboard(nil, donorStore, nil)(AppData{}, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
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

func TestReadDashboardForm(t *testing.T) {
	f := url.Values{
		"has-existing-donor-lpas": {"true"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	result := readDashboardForm(r)

	assert.True(t, result.hasExistingDonorLPAs)
}
