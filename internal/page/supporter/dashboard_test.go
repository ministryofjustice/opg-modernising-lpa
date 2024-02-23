package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?a=b", nil)

	keys := []dynamo.Key{{PK: "a", SK: "b"}}
	pagination := &search.Pagination{Total: 10}
	donors := []actor.DonorProvidedDetails{{LpaID: "abc"}}

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Query(r.Context(), search.QueryRequest{Page: 1, PageSize: 10}).
		Return(&search.QueryResponse{Keys: keys, Pagination: pagination}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetByKeys(r.Context(), keys).
		Return(donors, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:        testAppData,
			Donors:     donors,
			Query:      url.Values{"a": {"b"}},
			Pagination: pagination,
		}).
		Return(expectedError)

	err := Dashboard(template.Execute, donorStore, searchClient)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenSearchClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Query(r.Context(), search.QueryRequest{Page: 1, PageSize: 10}).
		Return(&search.QueryResponse{Keys: []dynamo.Key{}, Pagination: &search.Pagination{}}, expectedError)

	err := Dashboard(nil, nil, searchClient)(testAppData, w, r, nil)
	assert.Equal(t, expectedError, err)
}

func TestGetDashboardWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Query(r.Context(), search.QueryRequest{Page: 1, PageSize: 10}).
		Return(&search.QueryResponse{Keys: []dynamo.Key{}, Pagination: &search.Pagination{}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetByKeys(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := Dashboard(nil, donorStore, searchClient)(testAppData, w, r, nil)
	assert.Equal(t, expectedError, err)
}
