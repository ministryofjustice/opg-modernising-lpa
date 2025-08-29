package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMakeOrAddAnLPA(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{Donor: []dashboarddata.Actor{{}}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, makeOrAddAnLPAData{
			App:          appcontext.Data{},
			HasDonorLPAs: true,
		}).
		Return(nil)

	err := MakeOrAddAnLPA(template.Execute, nil, dashboardStore, nil)(appcontext.Data{}, w, r)

	assert.Nil(t, err)
}

func TestGetMakeOrAddAnLPAWhenDashboardStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(mock.Anything).
		Return(dashboarddata.Results{}, expectedError)

	err := MakeOrAddAnLPA(nil, nil, dashboardStore, nil)(appcontext.Data{Path: "/"}, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetMakeOrAddAnLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(mock.Anything).
		Return(dashboarddata.Results{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := MakeOrAddAnLPA(template.Execute, nil, dashboardStore, nil)(appcontext.Data{Path: "/"}, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostMakeOrAddAnLPA(t *testing.T) {
	testCases := map[string]struct {
		Results          dashboarddata.Results
		ExpectedRedirect donor.Path
	}{
		"no donor LPAs": {
			ExpectedRedirect: donor.PathYourName,
		},
		"with donor LPAs": {
			Results:          dashboarddata.Results{Donor: []dashboarddata.Actor{{}}},
			ExpectedRedirect: donor.PathMakeANewLPA,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			r.Header.Add("Content-Type", FormUrlEncoded)

			dashboardStore := newMockDashboardStore(t)
			dashboardStore.EXPECT().
				GetAll(r.Context()).
				Return(tc.Results, nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Create(r.Context()).
				Return(&donordata.Provided{LpaID: "lpa-id"}, nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendMetric(r.Context(), "lpa-id", event.CategoryFunnelStartRate, event.MeasureOnlineDonor).
				Return(nil)

			err := MakeOrAddAnLPA(nil, donorStore, dashboardStore, eventClient)(appcontext.Data{}, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostMakeOrAddAnLPAWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add("Content-Type", FormUrlEncoded)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Create(r.Context()).
		Return(nil, expectedError)

	err := MakeOrAddAnLPA(nil, donorStore, dashboardStore, nil)(appcontext.Data{}, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostMakeOrAddAnLPAWhenEventClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add("Content-Type", FormUrlEncoded)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(mock.Anything).
		Return(dashboarddata.Results{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Create(mock.Anything).
		Return(&donordata.Provided{}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := MakeOrAddAnLPA(nil, donorStore, dashboardStore, eventClient)(appcontext.Data{}, w, r)

	assert.ErrorIs(t, err, expectedError)
}
