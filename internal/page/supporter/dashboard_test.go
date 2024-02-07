package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &DashboardData{
			App: testAppData,
		}).
		Return(nil)

	err := Dashboard(template.Execute, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDashboard(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateLPA(r.Context(), "org-id").
		Return(&actor.DonorProvidedDetails{LpaID: "lpa-id"}, nil)

	err := Dashboard(nil, organisationStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostDashboardWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateLPA(r.Context(), "org-id").
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := Dashboard(nil, organisationStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
