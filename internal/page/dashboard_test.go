package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []LpaAndActorTasks{
		{Lpa: &lpadata.Lpa{LpaID: "123"}},
		{Lpa: &lpadata.Lpa{LpaID: "456"}},
	}

	certificateProviderLpas := []LpaAndActorTasks{{Lpa: &lpadata.Lpa{LpaID: "abc"}}}
	attorneyLpas := []LpaAndActorTasks{{Lpa: &lpadata.Lpa{LpaID: "def"}}}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(donorLpas, attorneyLpas, certificateProviderLpas, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:                     appcontext.Data{},
			UseTabs:                 true,
			DonorLpas:               donorLpas,
			AttorneyLpas:            attorneyLpas,
			CertificateProviderLpas: certificateProviderLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, nil, dashboardStore)(appcontext.Data{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardOnlyDonor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorLpas := []LpaAndActorTasks{
		{Lpa: &lpadata.Lpa{LpaID: "123"}},
		{Lpa: &lpadata.Lpa{LpaID: "456"}},
	}

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(donorLpas, nil, nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:       appcontext.Data{},
			DonorLpas: donorLpas,
		}).
		Return(nil)

	err := Dashboard(template.Execute, nil, dashboardStore)(appcontext.Data{}, w, r)
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

	err := Dashboard(nil, nil, dashboardStore)(appcontext.Data{}, w, r)
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

	err := Dashboard(template.Execute, nil, dashboardStore)(appcontext.Data{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostDashboard(t *testing.T) {
	testCases := map[string]struct {
		Form             url.Values
		ExpectedRedirect donor.Path
	}{
		"no donor LPAs": {
			ExpectedRedirect: donor.PathYourName,
		},
		"with donor LPAs": {
			Form:             url.Values{"has-existing-donor-lpas": {"true"}},
			ExpectedRedirect: donor.PathMakeANewLPA,
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
				Return(&donordata.Provided{LpaID: "lpa-id"}, nil)

			err := Dashboard(nil, donorStore, nil)(appcontext.Data{}, w, r)
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
		Return(&donordata.Provided{LpaID: "123"}, expectedError)

	err := Dashboard(nil, donorStore, nil)(appcontext.Data{}, w, r)
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
