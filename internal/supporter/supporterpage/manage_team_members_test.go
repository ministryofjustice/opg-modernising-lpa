package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetManageTeamMembers(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembers(r.Context()).
		Return([]*supporterdata.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}}, nil)
	memberStore.EXPECT().
		GetAll(r.Context()).
		Return([]*supporterdata.Member{{FirstNames: "c"}, {FirstNames: "d"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &manageTeamMembersData{
			App:            testAppData,
			Organisation:   &supporterdata.Organisation{ID: "org-id"},
			InvitedMembers: []*supporterdata.MemberInvite{{FirstNames: "a"}, {FirstNames: "b"}},
			Members:        []*supporterdata.Member{{FirstNames: "c"}, {FirstNames: "d"}},
		}).
		Return(nil)

	err := ManageTeamMembers(template.Execute, memberStore, nil, nil, "")(testAppData, w, r, &supporterdata.Organisation{ID: "org-id"}, nil)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetManageTeamMembersWhenMemberStoreErrors(t *testing.T) {
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
				Return([]*supporterdata.MemberInvite{}, tc.invitedMembersError)

			if tc.membersError != nil {
				memberStore.EXPECT().
					GetAll(mock.Anything).
					Return([]*supporterdata.Member{}, tc.membersError)
			}

			err := ManageTeamMembers(nil, memberStore, nil, nil, "")(testAppData, w, r, &supporterdata.Organisation{}, nil)

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
		Return([]*supporterdata.MemberInvite{}, nil)
	memberStore.EXPECT().
		GetAll(mock.Anything).
		Return([]*supporterdata.Member{{FirstNames: "c"}, {FirstNames: "d"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ManageTeamMembers(template.Execute, memberStore, nil, nil, "")(testAppData, w, r, &supporterdata.Organisation{}, nil)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostManageTeamMembers(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisation := &supporterdata.Organisation{Name: "My organisation", ID: "org-id"}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		DeleteMemberInvite(r.Context(), organisation.ID, "email@example.com").
		Return(nil)
	memberStore.EXPECT().
		CreateMemberInvite(r.Context(), organisation, "a", "b", "email@example.com", "abcde", supporterdata.PermissionAdmin).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), "email@example.com", notify.OrganisationMemberInviteEmail{
			OrganisationName:      "My organisation",
			InviterEmail:          "supporter@example.com",
			InviteCode:            "abcde",
			JoinAnOrganisationURL: "http://base" + page.PathSupporterStart.Format(),
		}).
		Return(nil)

	err := ManageTeamMembers(nil, memberStore, func(int) string { return "abcde" }, notifyClient, "http://base")(testOrgMemberAppData, w, r, organisation, nil)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathManageTeamMembers.Format()+"?inviteSent=email%40example.com", resp.Header.Get("Location"))
}

func TestPostManageTeamMembersWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"email":       {"not an email"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := ManageTeamMembers(nil, nil, func(int) string { return "abcde" }, nil, "http://base")(testAppData, w, r, &supporterdata.Organisation{ID: "org-id", Name: "My organisation"}, nil)

	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostManageTeamMembersWhenMemberStoreErrors(t *testing.T) {
	testcases := map[string]struct {
		deleteMembersError error
		createMemberInvite error
	}{
		"DeleteMemberInvite error": {
			deleteMembersError: expectedError,
		},
		"CreateMemberInvite error": {
			createMemberInvite: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"email":       {"email@example.com"},
				"first-names": {"a"},
				"last-name":   {"b"},
				"permission":  {"admin"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				DeleteMemberInvite(mock.Anything, mock.Anything, mock.Anything).
				Return(tc.deleteMembersError)

			if tc.createMemberInvite != nil {
				memberStore.EXPECT().
					CreateMemberInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tc.createMemberInvite)
			}

			err := ManageTeamMembers(nil, memberStore, func(int) string { return "abcde" }, nil, "")(testAppData, w, r, &supporterdata.Organisation{}, nil)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostManageTeamMembersWhenNotifyClientError(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		DeleteMemberInvite(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	memberStore.EXPECT().
		CreateMemberInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ManageTeamMembers(nil, memberStore, func(int) string { return "abcde" }, notifyClient, "")(testAppData, w, r, &supporterdata.Organisation{}, nil)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

}
