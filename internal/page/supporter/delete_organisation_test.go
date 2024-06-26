package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDeleteOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		CountWithQuery(r.Context(), search.CountWithQueryReq{MustNotExist: "RegisteredAt"}).
		Return(1, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &deleteOrganisationData{
			App:                testAppData,
			InProgressLPACount: 1,
		}).
		Return(nil)

	err := DeleteOrganisation(template.Execute, nil, nil, searchClient)(testAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteOrganisationNameWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		CountWithQuery(r.Context(), search.CountWithQueryReq{MustNotExist: "RegisteredAt"}).
		Return(0, expectedError)

	err := DeleteOrganisation(nil, nil, nil, searchClient)(testAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteOrganisationNameWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		CountWithQuery(r.Context(), search.CountWithQueryReq{MustNotExist: "RegisteredAt"}).
		Return(1, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(template.Execute, nil, nil, searchClient)(testAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDeleteOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		ClearLogin(r, w).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		SoftDelete(r.Context(), &actor.Organisation{}).
		Return(nil)

	err := DeleteOrganisation(nil, organisationStore, sessionStore, nil)(testOrgMemberAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.OrganisationDeleted.Format()+"?organisationName=My+organisation", resp.Header.Get("Location"))
}

func TestPostDeleteOrganisationNameWhenSessionStoreErrorsError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		SoftDelete(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		ClearLogin(mock.Anything, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(nil, organisationStore, sessionStore, nil)(testOrgMemberAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

}

func TestPostDeleteOrganisationNameWhenOrganisationStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		SoftDelete(mock.Anything, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(nil, organisationStore, nil, nil)(testOrgMemberAppData, w, r, &actor.Organisation{}, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
