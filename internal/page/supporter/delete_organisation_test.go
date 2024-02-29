package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDeleteOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &deleteOrganisationNameData{
			App: testAppData,
		}).
		Return(nil)

	err := DeleteOrganisation(template.Execute, nil, nil)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteOrganisationNameWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DeleteOrganisation(template.Execute, nil, nil)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDeleteOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := newMockSessionStore(t)

	session := sessions.NewSession(nil, "")
	session.Options = &sessions.Options{
		MaxAge: -1,
	}
	session.Values = map[any]any{}

	sessionStore.EXPECT().
		Get(r, "session").
		Return(&sessions.Session{
			Options: &sessions.Options{
				MaxAge: 86400,
			},
			Values: map[any]any{
				"some-session-data": "data",
			},
		}, nil)

	sessionStore.EXPECT().
		Save(r, w, session).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		SoftDelete(r.Context()).
		Return(nil)

	err := DeleteOrganisation(nil, organisationStore, sessionStore)(testOrgMemberAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.OrganisationDeleted.Format()+"?organisationName=My+organisation", resp.Header.Get("Location"))
}

func TestPostDeleteOrganisationNameWhenSessionStoreErrorsGetError(t *testing.T) {
	testcases := map[string]struct {
		getError  error
		saveError error
	}{
		"when get error": {
			getError: expectedError,
		},
		"when save error": {
			saveError: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(mock.Anything, mock.Anything).
				Return(&sessions.Session{Options: &sessions.Options{}}, tc.getError)

			if tc.saveError != nil {
				sessionStore.EXPECT().
					Save(mock.Anything, mock.Anything, mock.Anything).
					Return(tc.saveError)
			}

			err := DeleteOrganisation(nil, nil, sessionStore)(testOrgMemberAppData, w, r, &actor.Organisation{})
			resp := w.Result()

			assert.Error(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostDeleteOrganisationNameWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(&sessions.Session{
			Options: &sessions.Options{},
		}, nil)

	sessionStore.EXPECT().
		Save(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		SoftDelete(r.Context()).
		Return(expectedError)

	err := DeleteOrganisation(nil, organisationStore, sessionStore)(testOrgMemberAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
