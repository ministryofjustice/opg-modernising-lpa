package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	testcases := map[string]int{
		"/":           1,
		"/?page=5":    5,
		"/?page=what": 1,
	}

	for url, page := range testcases {
		t.Run(url, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, url, nil)

			keys := []dynamo.Keys{{PK: dynamo.LpaKey("a"), SK: dynamo.OrganisationKey("b")}}
			pagination := &search.Pagination{Total: 10}
			donors := []donordata.Provided{{LpaID: "abc"}}

			searchClient := newMockSearchClient(t)
			searchClient.EXPECT().
				Query(r.Context(), search.QueryRequest{Page: page, PageSize: 10}).
				Return(&search.QueryResponse{Keys: keys, Pagination: pagination}, nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				GetByKeys(r.Context(), keys).
				Return(donors, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &dashboardData{
					App:         testAppData,
					Donors:      donors,
					CurrentPage: page,
					Pagination:  pagination,
				}).
				Return(expectedError)

			err := Dashboard(template.Execute, donorStore, searchClient)(testAppData, w, r, nil, nil)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDashboardWhenSearchClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Query(r.Context(), search.QueryRequest{Page: 1, PageSize: 10}).
		Return(&search.QueryResponse{Keys: []dynamo.Keys{}, Pagination: &search.Pagination{}}, expectedError)

	err := Dashboard(nil, nil, searchClient)(testAppData, w, r, nil, nil)
	assert.Equal(t, expectedError, err)
}

func TestGetDashboardWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Query(r.Context(), search.QueryRequest{Page: 1, PageSize: 10}).
		Return(&search.QueryResponse{Keys: []dynamo.Keys{}, Pagination: &search.Pagination{}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetByKeys(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := Dashboard(nil, donorStore, searchClient)(testAppData, w, r, nil, nil)
	assert.Equal(t, expectedError, err)
}
