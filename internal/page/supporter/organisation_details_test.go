package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrganisationDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?a=b", nil)

	invitedMembers := []*actor.MemberInvite{
		{"PK", "SK", time.Now(), "org-id", "a@example.com", "a", "b", actor.None},
	}

	organisation := actor.Organisation{
		PK:   "PK",
		SK:   "SK",
		ID:   "org-id",
		Name: "Org Corp Ltd.",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &organisationDetailsData{
			App:            testAppData,
			Organisation:   &organisation,
			InvitedMembers: invitedMembers,
			Query:          url.Values{"a": {"b"}},
		}).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(r.Context()).
		Return(invitedMembers, nil)

	err := OrganisationDetails(template.Execute, organisationStore)(testAppData, w, r, &organisation)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOrganisationDetailsWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(r.Context()).
		Return([]*actor.MemberInvite{}, expectedError)

	err := OrganisationDetails(nil, organisationStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOrganisationDetailsWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMembers(mock.Anything).
		Return([]*actor.MemberInvite{}, nil)

	err := OrganisationDetails(template.Execute, organisationStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
