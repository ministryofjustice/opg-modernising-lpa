package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetManageTeamMembers(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?a=b", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(r.Context()).
		Return([]*actor.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &manageTeamMembersData{
			App:            testAppData,
			Query:          url.Values{"a": {"b"}},
			Organisation:   &actor.Organisation{ID: "org-id"},
			InvitedMembers: []*actor.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}},
		}).
		Return(nil)

	err := ManageTeamMembers(template.Execute, organisationStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetManageTeamMembersWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(mock.Anything).
		Return([]*actor.MemberInvite{}, expectedError)

	err := ManageTeamMembers(nil, organisationStore)(testAppData, w, r, &actor.Organisation{})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetManageTeamMembersWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(mock.Anything).
		Return([]*actor.MemberInvite{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ManageTeamMembers(template.Execute, organisationStore)(testAppData, w, r, &actor.Organisation{})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
