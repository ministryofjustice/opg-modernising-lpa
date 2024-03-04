package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDeleteOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		AllLPAs(r.Context()).
		Return([]actor.DonorProvidedDetails{{}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &deleteOrganisationData{
			App:                testAppData,
			InProgressLPACount: 1,
		}).
		Return(nil)

	err := DeleteOrganisation(template.Execute, organisationStore, nil)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteOrganisationNameWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		AllLPAs(mock.Anything).
		Return([]actor.DonorProvidedDetails{{}}, expectedError)

	err := DeleteOrganisation(nil, organisationStore, nil)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteOrganisationNameWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		AllLPAs(mock.Anything).
		Return([]actor.DonorProvidedDetails{{}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(template.Execute, organisationStore, nil)(testAppData, w, r, &actor.Organisation{})
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
		AllLPAs(r.Context()).
		Return([]actor.DonorProvidedDetails{{}}, nil)

	organisationStore.EXPECT().
		SoftDelete(r.Context()).
		Return(nil)

	err := DeleteOrganisation(nil, organisationStore, sessionStore)(testOrgMemberAppData, w, r, &actor.Organisation{})
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
		AllLPAs(mock.Anything).
		Return([]actor.DonorProvidedDetails{{}}, nil)
	organisationStore.EXPECT().
		SoftDelete(mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		ClearLogin(mock.Anything, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(nil, organisationStore, sessionStore)(testOrgMemberAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

}

func TestPostDeleteOrganisationNameWhenOrganisationStoreErrors(t *testing.T) {
	testcases := map[string]struct {
		allLPAsError    error
		softDeleteError error
	}{
		"when AllLPAs error": {
			allLPAsError: expectedError,
		},
		"when SoftDelete error": {
			softDeleteError: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			organisationStore := newMockOrganisationStore(t)
			organisationStore.EXPECT().
				AllLPAs(mock.Anything).
				Return([]actor.DonorProvidedDetails{}, tc.allLPAsError)

			if tc.softDeleteError != nil {
				organisationStore.EXPECT().
					SoftDelete(mock.Anything).
					Return(tc.softDeleteError)
			}

			err := DeleteOrganisation(nil, organisationStore, nil)(testOrgMemberAppData, w, r, &actor.Organisation{})
			resp := w.Result()

			assert.Error(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
