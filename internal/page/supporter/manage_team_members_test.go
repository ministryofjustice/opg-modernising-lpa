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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembers(r.Context()).
		Return([]*actor.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}}, nil)
	memberStore.EXPECT().
		GetAll(r.Context()).
		Return([]*actor.Member{{FirstNames: "c"}, {FirstNames: "d"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &manageTeamMembersData{
			App:            testAppData,
			Query:          url.Values{"a": {"b"}},
			Organisation:   &actor.Organisation{ID: "org-id"},
			InvitedMembers: []*actor.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}},
			Members:        []*actor.Member{{FirstNames: "c"}, {FirstNames: "d"}},
		}).
		Return(nil)

	err := ManageTeamMembers(template.Execute, memberStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetManageTeamMembersWhenOrganisationStoreErrors(t *testing.T) {
	testcases := map[string]struct {
		invitedMembersError error
		membersError        error
	}{
		"InvitedMembers error": {
			invitedMembersError: expectedError,
		},
		"GetAll error": {
			membersError: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				InvitedMembers(mock.Anything).
				Return([]*actor.MemberInvite{}, tc.invitedMembersError)

			if tc.membersError != nil {
				memberStore.EXPECT().
					GetAll(mock.Anything).
					Return([]*actor.Member{}, tc.membersError)
			}

			err := ManageTeamMembers(nil, memberStore)(testAppData, w, r, &actor.Organisation{})

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetManageTeamMembersWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembers(mock.Anything).
		Return([]*actor.MemberInvite{}, nil)
	memberStore.EXPECT().
		GetAll(mock.Anything).
		Return([]*actor.Member{{FirstNames: "c"}, {FirstNames: "d"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ManageTeamMembers(template.Execute, memberStore)(testAppData, w, r, &actor.Organisation{})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
