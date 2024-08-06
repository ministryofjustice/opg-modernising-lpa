package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
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

	err := DeleteOrganisation(nil, template.Execute, nil, nil, searchClient)(testAppData, w, r, &supporterdata.Organisation{}, nil)
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

	err := DeleteOrganisation(nil, nil, nil, nil, searchClient)(testAppData, w, r, &supporterdata.Organisation{}, nil)
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

	err := DeleteOrganisation(nil, template.Execute, nil, nil, searchClient)(testAppData, w, r, &supporterdata.Organisation{}, nil)
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
		SoftDelete(r.Context(), &supporterdata.Organisation{}).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "organisation deleted")

	err := DeleteOrganisation(logger, nil, organisationStore, sessionStore, nil)(testOrgMemberAppData, w, r, &supporterdata.Organisation{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterOrganisationDeleted.Format()+"?organisationName=My+organisation", resp.Header.Get("Location"))
}

func TestPostDeleteOrganisationNameWhenSessionStoreErrors(t *testing.T) {
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

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything)

	err := DeleteOrganisation(logger, nil, organisationStore, sessionStore, nil)(testOrgMemberAppData, w, r, &supporterdata.Organisation{}, nil)
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

	err := DeleteOrganisation(nil, nil, organisationStore, nil, nil)(testOrgMemberAppData, w, r, &supporterdata.Organisation{}, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
