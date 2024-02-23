package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donors := []actor.DonorProvidedDetails{{LpaID: "abc"}}

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		AllLPAs(r.Context()).
		Return(donors, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dashboardData{
			App:    testAppData,
			Donors: donors,
		}).
		Return(expectedError)

	err := Dashboard(template.Execute, organisationStore)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDashboardWhenOrganisationStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		AllLPAs(r.Context()).
		Return(nil, expectedError)

	err := Dashboard(nil, organisationStore)(testAppData, w, r, nil)
	assert.Equal(t, expectedError, err)
}
